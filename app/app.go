package app

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
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

func InitDB() (*sql.DB, error) {
	connString, present := os.LookupEnv("CONNECTION_STRING")
	if !present {
		panic("shouldn't happen: connection string missing; should have been checked in TestMain")
	}
	return sql.Open("postgres", connString)
}
