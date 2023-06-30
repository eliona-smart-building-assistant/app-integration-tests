package docker

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/eliona-smart-building-assistant/app-integration-tests/app"

	_ "github.com/lib/pq"
)

// Assuming Dockerfile is present in the current directory
var (
	dockerBuildCmd = []string{"build", ".", "-t", "go-app-test"}
	dockerRunCmd   = "run --name go-app-test-container --network eliona-mock-network -d -i -p 3039:3039 -e 'API_ENDPOINT=%s' -e 'API_TOKEN=%s' -e 'CONNECTION_STRING=%s' -e 'LOG_LEVEL=info' -e 'API_SERVER_PORT=3039' go-app-test"
	dockerLogsCmd  = []string{"logs", "-f", "go-app-test-container"}
	dockerStopCmd  = []string{"stop", "go-app-test-container"}
	dockerRmCmd    = []string{"rm", "go-app-test-container"}
)

var (
	appLocation string
	metadata    app.Metadata

	apiEndpoint string
	apiToken    string
	connString  string
)

func RunApp(m *testing.M) {
	StartApp()
	code := m.Run()
	StopApp()
	os.Exit(code)
}

func StartApp() {
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

	var err error
	metadata, err = app.GetMetadata()
	if err != nil {
		fmt.Printf("getting metadata: %v", err)
		os.Exit(1)
	}

	db, err := initDB()
	if err != nil {
		fmt.Printf("initializing db: %v", err)
		os.Exit(1)
	}

	if err := resetDB(db); err != nil {
		fmt.Printf("resetting db: %v", err)
		os.Exit(1)
	}

	if err := addAppToStore(db); err != nil {
		fmt.Printf("adding app to app store: %v", err)
		os.Exit(1)
	}

	// Build and run docker image
	fmt.Println("Building the image...")
	{
		out, err := exec.Command("docker", dockerBuildCmd...).CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to build docker image: %s\n%s", err, out)
			os.Exit(1)
		}
	}

	{
		cmdStr := fmt.Sprintf(dockerRunCmd, apiEndpoint, apiToken, connString)
		cmd := exec.Command("bash", "-c", "docker "+cmdStr)
		if runtime.GOOS == "windows" {
			cmd = exec.Command("cmd", "/C", cmdStr)
		}
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to start docker container: %s\n%s", err, out)
			os.Exit(1)
		}
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

func checkEnvVars() error {
	var present bool

	apiEndpoint, present = os.LookupEnv("API_ENDPOINT")
	if !present {
		return errors.New("API_ENDPOINT variable not defined.")
	}

	apiToken, present = os.LookupEnv("API_TOKEN")
	if !present {
		return errors.New("API_TOKEN variable not defined.")
	}

	connString, present = os.LookupEnv("CONNECTION_STRING")
	if !present {
		return errors.New("CONNECTION_STRING variable not defined.")
	}
	return nil
}

func resetDB(db *sql.DB) error {
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

func initDB() (*sql.DB, error) {
	connString, present := os.LookupEnv("CONNECTION_STRING")
	if !present {
		panic("shouldn't happen: connection string missing; should have been checked in TestMain")
	}
	return sql.Open("postgres", connString)
}

func addAppToStore(db *sql.DB) error {
	iconFile, err := os.Open("icon")
	if err != nil {
		return fmt.Errorf("opening icon file: %s", err)
	}
	defer iconFile.Close()
	iconData, err := io.ReadAll(iconFile)
	if err != nil {
		return fmt.Errorf("reading icon file: %s", err)
	}

	metadataFile, err := os.Open("metadata.json")
	if err != nil {
		return fmt.Errorf("opening metadata file: %s", err)
	}
	defer metadataFile.Close()
	metadataData, err := io.ReadAll(metadataFile)
	if err != nil {
		return fmt.Errorf("reading metadata file: %s", err)
	}

	if _, err := db.Exec(`
		UPDATE eliona_store
		SET metadata = $1, icon = $2
		WHERE app_name = $3`, string(metadataData), string(iconData), metadata.Name); err != nil {
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
