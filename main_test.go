package integration_test

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
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

var (
	appLocation string
	db          *sql.DB
	metadata    Metadata

	apiEndpoint string
	apiToken    string
	connString  string
)

func TestMain(m *testing.M) {
	handleFlags()
	if err := os.Chdir(appLocation); err != nil {
		fmt.Printf("chdir to %s: %v", appLocation, err)
		os.Exit(1)
	}

	if err := checkEnvVars(); err != nil {
		fmt.Printf("checking environment variables: %v", err)
		os.Exit(1)
	}

	md, err := getMetadata()
	if err != nil {
		fmt.Printf("getting metadata: %v", err)
		os.Exit(1)
	}
	if md == nil {
		panic("shouldn't happen: metadata is nil")
	}
	metadata = *md

	if err := resetDB(); err != nil {
		fmt.Printf("resetting db: %v", err)
		os.Exit(1)
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

func handleFlags() {
	flag.StringVar(&appLocation, "app", "", "Path to app")
	flag.Parse()

	if appLocation == "" {
		fmt.Println("App path must be provided. Use -app argument.")
		os.Exit(1)
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

func resetDB() error {
	db, err := initDB()
	if err != nil {
		return fmt.Errorf("opening database connection: %s", err)
	}
	defer db.Close()

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

type Metadata struct {
	Name                   string            `json:"name"`
	ElionaMinVersion       string            `json:"elionaMinVersion"`
	DisplayName            map[string]string `json:"displayName"`
	Description            map[string]string `json:"description"`
	DashboardTemplateNames []string          `json:"dashboardTemplateNames"`
	ApiUrl                 string            `json:"apiUrl"`
	ApiSpecificationPath   string            `json:"apiSpecificationPath"`
	DocumentationUrl       string            `json:"documentationUrl"`
	UseEnvironment         []string          `json:"useEnvironment"`
}

func getMetadata() (*Metadata, error) {
	file, err := os.Open("metadata.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata.json: %w", err)
	}
	defer file.Close()

	var metadata Metadata
	err = json.NewDecoder(file).Decode(&metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to decode metadata.json: %w", err)
	}

	return &metadata, nil
}
