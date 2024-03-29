//  This file is part of the eliona project.
//  Copyright © 2022 LEICOM iTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/eliona-smart-building-assistant/app-integration-tests/app"
	eapp "github.com/eliona-smart-building-assistant/go-eliona/app"
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
	if app.StartMode() == app.StartModeDocker {
		assert.NotEmpty(t, versionResponse.Commit, "Commit field is not empty")
		assert.NotEmpty(t, versionResponse.Timestamp, "Timestamp field is not empty")
	}
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

func getMetadata(t *testing.T) eapp.Metadata {
	metadata, _, err := eapp.GetMetadata()
	require.NoError(t, err, "Getting metadata successful")
	return metadata
}

func getUrl(t *testing.T, url string) *http.Response {
	resp, err := http.Get(url)
	require.NoErrorf(t, err, "%s should be accessible", url)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Expected status OK")
	return resp
}
