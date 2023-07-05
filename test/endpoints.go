package test

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
	metadata := getMetadata(t)
	resp := getUrl(t, fmt.Sprintf("http://localhost:3039/%s/version", metadata.ApiUrl))
	defer resp.Body.Close()

	versionResponse := decodeResponse[VersionResponse](t, resp)
	assert.NotEmpty(t, versionResponse.Commit, "Commit field is not empty")
	assert.NotEmpty(t, versionResponse.Timestamp, "Timestamp field is not empty")
}

func APISpecEndpointExists(t *testing.T) {
	metadata := getMetadata(t)
	resp := getUrl(t, fmt.Sprintf("http://localhost:3039/%s/%s", metadata.ApiUrl, metadata.ApiSpecificationPath))
	defer resp.Body.Close()

	_ = decodeResponse[any](t, resp)
}

func decodeResponse[T any](t *testing.T, resp *http.Response) T {
	var decoded T
	err := json.NewDecoder(resp.Body).Decode(&decoded)
	require.NoError(t, err, "Decoding response body")
	return decoded
}

func getMetadata(t *testing.T) app.Metadata {
	metadata, _, err := app.GetMetadata()
	require.NoError(t, err, "Getting metadata successful")
	return metadata
}

func getUrl(t *testing.T, url string) *http.Response {
	resp, err := http.Get(url)
	require.NoErrorf(t, err, "%s should be accessible", url)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK")
	return resp
}
