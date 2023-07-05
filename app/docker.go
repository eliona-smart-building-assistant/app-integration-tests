package app

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// Assuming Dockerfile is present in the current directory
var (
	dockerBuildCmd = []string{"build", ".",
		"-t", "go-app-test"}
	dockerRunCmd = []string{"run",
		"--name", "go-app-test-container",
		"--network", "eliona-mock-network",
		"-d",
		"-i",
		"-p", "3039:3000",
		"-e", "API_ENDPOINT=$API_ENDPOINT",
		"-e", "API_TOKEN=$API_TOKEN",
		"-e", "CONNECTION_STRING=$CONNECTION_STRING",
		"-e", "LOG_LEVEL=info",
		"go-app-test"}
	dockerLogsCmd = []string{"logs", "-f", "go-app-test-container"}
	dockerStopCmd = []string{"stop", "go-app-test-container"}
	dockerRmCmd   = []string{"rm", "go-app-test-container"}
)

func startAppContainer() {
	out, err := exec.Command("docker", dockerRmCmd...).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to remove docker image: %s\n%s", err, out)
	}

	// Build and run docker image
	fmt.Println("Building the image...")
	out, err = exec.Command("docker", dockerBuildCmd...).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to build docker image: %s\n%s", err, out)
		os.Exit(1)
	}

	out, err = exec.Command("docker", expandEnvInArray(dockerRunCmd...)...).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to start docker container: %s\n%s", err, out)
		os.Exit(1)
	}

	go monitorDockerLogs()

	if err := waitForContainerReady(); err != nil {
		fmt.Printf("waiting for container to get ready: %v\n", err)
		os.Exit(1)
	}
}

func stopAppContainer() {
	// Cool down period to notice any errors occurring later after running tests.
	time.Sleep(time.Second * 1)
	teardownDocker()
}

func expandEnvInArray(arr ...string) []string {
	result := make([]string, len(arr))
	for i, str := range arr {
		result[i] = os.ExpandEnv(str)
	}
	return result
}

func teardownDocker() {
	out, err := exec.Command("docker", dockerStopCmd...).CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to stop docker container: %s\n%s", err, out)
		os.Exit(1)
	}
}

func waitForContainerReady() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	fmt.Println("Waiting for the container to get ready...")

	metadata, _, err := GetMetadata()
	if err != nil {
		return fmt.Errorf("getting metadata: %s", err)
	}

	url := fmt.Sprintf("http://localhost:3039/%s/version", metadata.ApiUrl)
	client := &http.Client{Timeout: 100 * time.Millisecond}
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("container did not become ready in the specified time")
		case <-time.After(time.Millisecond * 200):
			resp, err := client.Get(url)
			if err != nil {
				continue
			}
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
}

func monitorDockerLogs() {
	defer teardownDocker()

	logCmd := exec.Command("docker", dockerLogsCmd...)
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
