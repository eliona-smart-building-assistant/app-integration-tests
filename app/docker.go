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

package app

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// We tried using the Docker SDK, but the version of SDK is dependent on version
// of Docker daemon. Therefore, there were often compatibility problems.
// If this will be resolved in the future, it would be nice to use the SDK.

// Assuming Dockerfile is present in the current directory
var (
	dockerBuildCmd = []string{"build", ".",
		"-t", "go-app-test"}
	dockerRunCmd = []string{"run",
		"--name", "go-app-test-container",
		"-d",
		"-i",
		"-p", "3039:3000",
		"-e", "API_ENDPOINT=$API_ENDPOINT",
		"-e", "API_TOKEN=$API_TOKEN",
		"-e", "CONNECTION_STRING=$CONNECTION_STRING",
		"-e", "LOG_LEVEL=info",
		"--add-host", "host.docker.internal:host-gateway",
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

	if err := waitForAppReady(); err != nil {
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
