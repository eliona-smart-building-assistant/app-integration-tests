package assert

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/eliona-smart-building-assistant/app-integration-tests/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type VersionResponse struct {
	Commit    string `json:"commit"`
	Timestamp string `json:"timestamp"`
}

func VersionEndpointExists(t *testing.T) {
	metadata, err := app.GetMetadata()
	require.NoError(t, err, "Getting metadata successful")

	url := fmt.Sprintf("http://localhost:3039/%s/version", metadata.ApiUrl)
	resp, err := http.Get(url)
	require.NoError(t, err, "/version endpoint shouldn't report any error")
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK")

	var versionResponse VersionResponse
	err = json.NewDecoder(resp.Body).Decode(&versionResponse)
	require.NoError(t, err, "Decoding response body")

	assert.NotEmpty(t, versionResponse.Commit, "Commit field is not empty")
	assert.NotEmpty(t, versionResponse.Timestamp, "Timestamp field is not empty")
}

func APISpecEndpointExists(t *testing.T) {
	metadata, err := app.GetMetadata()
	require.NoError(t, err, "Getting metadata successful")

	url := "http://localhost:3039/" + metadata.ApiUrl + metadata.ApiSpecificationPath
	resp, err := http.Get(url)
	require.NoError(t, err, metadata.ApiSpecificationPath+" endpoint shouldn't report any error")

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK")

	var apiSpecResponse interface{}
	err = json.NewDecoder(resp.Body).Decode(&apiSpecResponse)
	require.NoError(t, err, "Decoding response body")
}
