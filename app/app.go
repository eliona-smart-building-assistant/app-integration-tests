package app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"testing"
	"time"
)

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

var (
	appLocation string
)

func RunApp(m *testing.M) {
	resetDB()
	startApp()
	code := m.Run()
	stoppApp()
	os.Exit(code)
}

func handleFlags() {
	flag.StringVar(&appLocation, "app", "", "Path to app")
	flag.Parse()

	if appLocation == "" {
		appLocation = "."
	}
}

func startApp() {
	if err := checkEnvVars(); err != nil {
		fmt.Printf("checking environment variables: %v", err)
		os.Exit(1)
	}

	handleFlags()
	if err := os.Chdir(appLocation); err != nil {
		fmt.Printf("chdir to %s: %v", appLocation, err)
		os.Exit(1)
	}
	startAppContainer()
}

func stoppApp() {
	stopAppContainer()
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

func GetMetadata() (Metadata, []byte, error) {
	file, err := os.Open("metadata.json")
	if err != nil {
		return Metadata{}, nil, fmt.Errorf("failed to open metadata.json: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return Metadata{}, data, fmt.Errorf("reading metadata file: %s", err)
	}

	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return Metadata{}, data, fmt.Errorf("failed unmarhalling metadata.json: %w", err)
	}

	return metadata, data, nil
}

func resetDB() {
	metadata, _, err := GetMetadata()
	if err != nil {
		fmt.Printf("getting metadata: %s", err)
		os.Exit(1)
	}

	db, err := GetDB()
	if err != nil {
		fmt.Printf("initializing db: %v", err)
		os.Exit(1)
	}

	sqlScript, err := os.ReadFile("reset.sql")
	if err != nil {
		fmt.Printf("reading SQL script: %s", err)
		os.Exit(1)
	}

	_, err = db.Exec(string(sqlScript))
	if err != nil {
		fmt.Printf("executing SQL script: %s", err)
		os.Exit(1)
	}

	row := db.QueryRow(`
		SELECT initialized_at
		FROM public.eliona_app
		WHERE app_name = $1;
	`, metadata.Name)
	var initialized *time.Time
	if err = row.Scan(&initialized); err != nil {
		fmt.Printf("executing SELECT statement: %s", err)
		os.Exit(1)
	}

	// Check if the script reset the initialization state of app
	if initialized != nil {
		fmt.Printf("unexpected result from SELECT statement: got %v, want nil", initialized)
		os.Exit(1)
	}
}

func GetDB() (*sql.DB, error) {
	connString, present := os.LookupEnv("CONNECTION_STRING")
	if !present {
		panic("shouldn't happen: connection string missing; should have been checked in TestMain")
	}
	return sql.Open("postgres", connString)
}
