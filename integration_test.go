package integration_test

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"
)

// Assuming Dockerfile is present in the current directory
const (
	dockerBuildCmd = "docker build -t go-app-test %s"
	dockerRunCmd   = "docker run --rm --name go-app-container -d -i -p 3039:3039 -e 'API_ENDPOINT=%s' -e 'API_TOKEN=%s' -e 'CONNECTION_STRING=%s' -e 'LOG_LEVEL=info' -e 'API_SERVER_PORT=3039' go-app-test"
	dockerStopCmd  = "docker stop go-app-container"
)

func TestMain(m *testing.M) {
	// TODO: Get location as param.
	dockerfilePath := "../kontakt-io-app/"

	apiEndpoint, present := os.LookupEnv("API_ENDPOINT")
	if !present {
		fmt.Printf("API_ENDPOINT variable not defined.")
		return
	}
	apiToken, present := os.LookupEnv("API_TOKEN")
	if !present {
		fmt.Printf("API_TOKEN variable not defined.")
		return
	}
	connString, present := os.LookupEnv("CONNECTION_STRING")
	if !present {
		fmt.Printf("CONNECTION_STRING variable not defined.")
		return
	}

	// Build and run docker image
	{
		cmd := fmt.Sprintf(dockerBuildCmd, dockerfilePath)
		out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to build docker image: %s\n%s", err, out)
			return
		}
	}

	{
		cmd := fmt.Sprintf(dockerRunCmd, apiEndpoint, apiToken, connString)
		out, err := exec.Command("/bin/sh", "-c", cmd).CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to start docker container: %s\n%s", err, out)
			return
		}
	}

	// Wait for the server to start
	time.Sleep(time.Second * 5)

	// Run the tests
	result := m.Run()

	// Stop docker container
	{
		out, err := exec.Command("/bin/sh", "-c", dockerStopCmd).CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to stop docker container: %s\n%s", err, out)
		}
	}
	os.Exit(result)
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
}
