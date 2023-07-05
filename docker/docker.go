package docker

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/eliona-smart-building-assistant/app-integration-tests/app"

	_ "github.com/lib/pq"
)

// Assuming Dockerfile is present in the current directory
var (
	dockerBuildCmd = []string{"build", ".",
		"-t", "go-app-test"}
	dockerRunCmd = []string{"run",
		"--name", "go-app-test-container",
		"--network", "eliona-mock-network",
		"-d",
		"-i",
		"-p", "3039:3000",
		"-e", "API_ENDPOINT=$API_ENDPOINT",
		"-e", "API_TOKEN=$API_TOKEN",
		"-e", "CONNECTION_STRING=$CONNECTION_STRING",
		"-e", "LOG_LEVEL=info",
		"go-app-test"}
	dockerLogsCmd = []string{"logs", "-f", "go-app-test-container"}
	dockerStopCmd = []string{"stop", "go-app-test-container"}
	dockerRmCmd   = []string{"rm", "go-app-test-container"}
)

var (
	appLocation string
)

func RunApp(m *testing.M) {
	startApp()
	code := m.Run()
	StopApp()
	os.Exit(code)
}

func startApp() {
	{
		out, err := exec.Command("docker", dockerRmCmd...).CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to remove docker image: %s\n%s", err, out)
		}
	}

	handleFlags()
	if err := os.Chdir(appLocation); err != nil {
		fmt.Printf("chdir to %s: %v", appLocation, err)
		os.Exit(1)
	}

	if err := checkEnvVars(); err != nil {
		fmt.Printf("checking environment variables: %v", err)
		os.Exit(1)
	}

	if err := resetDB(); err != nil {
		fmt.Printf("resetting db: %v", err)
		os.Exit(1)
	}

	// Build and run docker image
	fmt.Println("Building the image...")
	out, err := exec.Command("docker", dockerBuildCmd...).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to build docker image: %s\n%s", err, out)
		os.Exit(1)
	}

	out, err = exec.Command("docker", expandEnvInArray(dockerRunCmd...)...).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to start docker container: %s\n%s", err, out)
		os.Exit(1)
	}

	go monitorLogs()

	if err := waitForContainerReady(); err != nil {
		fmt.Printf("waiting for container to get ready: %v\n", err)
		os.Exit(1)
	}
}

func StopApp() {
	// Cool down period to notice any errors occurring later after running tests.
	time.Sleep(time.Second * 1)
	teardown()
}

func handleFlags() {
	flag.StringVar(&appLocation, "app", "", "Path to app")
	flag.Parse()

	if appLocation == "" {
		appLocation = "."
	}
}

func expandEnvInArray(arr ...string) []string {
	result := make([]string, len(arr))
	for i, str := range arr {
		result[i] = os.ExpandEnv(str)
	}
	return result
}

func checkEnvVars() error {
	var present bool

	_, present = os.LookupEnv("API_ENDPOINT")
	if !present {
		return errors.New("API_ENDPOINT variable not defined.")
	}

	_, present = os.LookupEnv("API_TOKEN")
	if !present {
		return errors.New("API_TOKEN variable not defined.")
	}

	_, present = os.LookupEnv("CONNECTION_STRING")
	if !present {
		return errors.New("CONNECTION_STRING variable not defined.")
	}
	return nil
}

func resetDB() error {
	metadata, _, err := app.GetMetadata()
	if err != nil {
		return fmt.Errorf("getting metadata: %s", err)
	}

	db, err := app.InitDB()
	if err != nil {
		fmt.Printf("initializing db: %v", err)
		os.Exit(1)
	}

	sqlScript, err := os.ReadFile("reset.sql")
	if err != nil {
		return fmt.Errorf("reading SQL script: %s", err)
	}

	_, err = db.Exec(string(sqlScript))
	if err != nil {
		return fmt.Errorf("executing SQL script: %s", err)
	}

	row := db.QueryRow(`
		SELECT initialized_at
		FROM public.eliona_app
		WHERE app_name = $1;
	`, metadata.Name)
	var initialized *time.Time
	if err = row.Scan(&initialized); err != nil {
		return fmt.Errorf("executing SELECT statement: %s\n", err)
	}

	// Check if the script reset the initialization state of app
	if initialized != nil {
		return fmt.Errorf("unexpected result from SELECT statement: got %v, want nil\n", initialized)
	}
	return nil
}

func teardown() {
	out, err := exec.Command("docker", dockerStopCmd...).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to stop docker container: %s\n%s", err, out)
		os.Exit(1)
	}
}

func waitForContainerReady() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	fmt.Println("Waiting for the container to get ready...")

	metadata, _, err := app.GetMetadata()
	if err != nil {
		return fmt.Errorf("getting metadata: %s", err)
	}

	url := fmt.Sprintf("http://localhost:3039/%s/version", metadata.ApiUrl)
	client := &http.Client{Timeout: 100 * time.Millisecond}
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("container did not become ready in the specified time")
		case <-time.After(time.Millisecond * 200):
			resp, err := client.Get(url)
			if err != nil {
				continue
			}
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
}

func monitorLogs() {
	defer teardown()

	logCmd := exec.Command("docker", dockerLogsCmd...)
	// All output is written to stderr.
	stderr, err := logCmd.StderrPipe()
	if err != nil {
		panic(fmt.Sprintf("Log stderr pipe: %v\n", err))
	}
	if err := logCmd.Start(); err != nil {
		panic(fmt.Sprintf("Starting log: %v\n", err))
	}

	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "FATAL") || strings.HasPrefix(line, "ERROR") {
			panic(fmt.Sprintf("Container log error: %s\n", line))
		}
	}
}
