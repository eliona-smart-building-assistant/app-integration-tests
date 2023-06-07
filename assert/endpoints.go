package assert

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/eliona-smart-building-assistant/app-integration-tests/app"
)

type VersionResponse struct {
	Commit    string `json:"commit"`
	Timestamp string `json:"timestamp"`
}

func VersionEndpointExists(t *testing.T, metadata app.Metadata) bool {
	url := fmt.Sprintf("http://localhost:3039/%s/version", metadata.ApiUrl)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to send request to /version: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK but got %d", resp.StatusCode)
	}

	var versionResponse VersionResponse
	err = json.NewDecoder(resp.Body).Decode(&versionResponse)
	if err != nil {
		t.Fatalf("Failed to decode response body: %s", err)
	}
	if versionResponse.Commit == "" {
		t.Error("Commit field is empty")
	}
	if versionResponse.Timestamp == "" {
		t.Error("Timestamp field is empty")
	}
	return true
}

func APISpecEndpointExists(t *testing.T, metadata app.Metadata) {
	url := "http://localhost:3039/" + metadata.ApiUrl + metadata.ApiSpecificationPath
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("Failed to send request to %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK but got %d", resp.StatusCode)
	}

	var apiSpecResponse interface{}
	err = json.NewDecoder(resp.Body).Decode(&apiSpecResponse)
	if err != nil {
		t.Fatalf("Failed to decode response body: %s", err)
	}
}
