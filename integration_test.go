package integration_test

import (
	"bufio"
	"database/sql"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

// Assuming Dockerfile is present in the current directory
const (
	dockerBuildCmd = "docker build . -t go-app-test"
	dockerRunCmd   = "docker run --rm --name go-app-container -d -i -p 3039:3039 -e 'API_ENDPOINT=%s' -e 'API_TOKEN=%s' -e 'CONNECTION_STRING=%s' -e 'LOG_LEVEL=info' -e 'API_SERVER_PORT=3039' go-app-test"
	dockerLogsCmd  = "docker logs -f go-app-container"
	dockerStopCmd  = "docker stop go-app-container"
)

var appLocation string

func TestMain(m *testing.M) {
	flag.StringVar(&appLocation, "app", "", "Path to app")
	flag.Parse()

	if appLocation == "" {
		fmt.Println("App path must be provided. Use -app argument.")
		os.Exit(1)
	}
	os.Chdir(appLocation)

	apiEndpoint, present := os.LookupEnv("API_ENDPOINT")
	if !present {
		fmt.Printf("API_ENDPOINT variable not defined.")
		os.Exit(1)
	}
	apiToken, present := os.LookupEnv("API_TOKEN")
	if !present {
		fmt.Printf("API_TOKEN variable not defined.")
		os.Exit(1)
	}
	connString, present := os.LookupEnv("CONNECTION_STRING")
	if !present {
		fmt.Printf("CONNECTION_STRING variable not defined.")
		os.Exit(1)
	}

	if err := resetDB(); err != nil {
		fmt.Printf("resetting db: %v", err)
	}

	// Build and run docker image
	{
		out, err := exec.Command("/bin/sh", "-c", dockerBuildCmd).CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to build docker image: %s\n%s", err, out)
			os.Exit(1)
		}
	}
	{
		cmd := fmt.Sprintf(dockerRunCmd, apiEndpoint, apiToken, connString)
		out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to start docker container: %s\n%s", err, out)
			os.Exit(1)
		}
	}

	go monitorLogs()

	// Wait for the server to start
	time.Sleep(time.Second * 5)

	// Run the tests
	result := m.Run()

	// Cool down period to notice any errors occuring later after running tests.
	time.Sleep(time.Second * 1)
	teardown()
	os.Exit(result)
}

func resetDB() error {
	db, err := initDB()
	if err != nil {
		return fmt.Errorf("opening database connection: %s", err)
	}
	defer db.Close()

	sqlScript, err := ioutil.ReadFile("reset.sql")
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
		WHERE app_name = 'kontakt-io'
	`)
	var initialized *time.Time
	err = row.Scan(&initialized)
	if err != nil {
		return fmt.Errorf("executing SELECT statement: %s", err)
	}

	// Check if the script reset the initialization state of app
	if initialized != nil {
		return fmt.Errorf("unexpected result from SELECT statement: got %v, want nil", initialized)
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

func teardown() {
	out, err := exec.Command("/bin/sh", "-c", dockerStopCmd).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to stop docker container: %s\n%s", err, out)
		os.Exit(1)
	}
}

func monitorLogs() {
	defer teardown()

	logCmd := exec.Command("/bin/sh", "-c", dockerLogsCmd)
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

func TestVersionEndpoint(t *testing.T) {
	// Test the /version endpoint
	resp, err := http.Get("http://localhost:3039/v1/version")
	if err != nil {
		t.Fatalf("Failed to send request to /version: %s", err)
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK but got %d", resp.StatusCode)
	}

	// TODO: Check if any content.
}
