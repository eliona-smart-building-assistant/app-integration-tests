package integration_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

type VersionResponse struct {
	Commit    string `json:"commit"`
	Timestamp string `json:"timestamp"`
}

func TestVersionEndpoint(t *testing.T) {
	resp, err := http.Get("http://localhost:3039/v1/version")
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
}
