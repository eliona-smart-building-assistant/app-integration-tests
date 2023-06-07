package app

import (
	"encoding/json"
	"fmt"
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

func GetMetadata() (Metadata, error) {
	file, err := os.Open("metadata.json")
	if err != nil {
		return Metadata{}, fmt.Errorf("failed to open metadata.json: %w", err)
	}
	defer file.Close()

	var metadata Metadata
	if err := json.NewDecoder(file).Decode(&metadata); err != nil {
		return Metadata{}, fmt.Errorf("failed to decode metadata.json: %w", err)
	}

	return metadata, nil
}
