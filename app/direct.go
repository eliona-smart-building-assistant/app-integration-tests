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
	"github.com/eliona-smart-building-assistant/go-eliona/app"
	_ "github.com/lib/pq"
	"io"
	"os"
	"os/exec"
	"strings"
)

// Assuming Dockerfile is present in the current directory
var (
	goRunCmdParams = []string{"run", "."}
	goRunCmd       *exec.Cmd
)

func startAppDirectly() {
	metadata, _, err := app.GetMetadata()
	if err != nil {
		fmt.Printf("getting metadata: %s", err)
		os.Exit(1)
	}

	goRunCmd = exec.Command("go", goRunCmdParams...)
	goRunCmd.Env = os.Environ()
	goRunCmd.Env = append(goRunCmd.Env, fmt.Sprintf("APPNAME=%s", metadata.Name))
	goRunCmd.Env = append(goRunCmd.Env, "API_SERVER_PORT=3039")

	// Create pipes to capture stdout and stderr
	stdoutPipe, err := goRunCmd.StdoutPipe()
	if err != nil {
		fmt.Println("Failed to create stdout pipe:", err)
		return
	}

	stderrPipe, err := goRunCmd.StderrPipe()
	if err != nil {
		fmt.Println("Failed to create stderr pipe:", err)
		return
	}

	// Start the command
	if err := goRunCmd.Start(); err != nil {
		fmt.Printf("Failed to start app directly: %s", err)
		os.Exit(1)
	}

	go monitorDirectOutput(stdoutPipe)
	go monitorDirectOutput(stderrPipe)

	if err := waitForAppReady(); err != nil {
		fmt.Printf("waiting for command to get ready: %v\n", err)
		os.Exit(1)
	}

}

func stopAppDirectly() {
	err := goRunCmd.Process.Kill()
	if err != nil {
		fmt.Println("Failed to send SIGKILL:", err)
		os.Exit(1)
	}
}

func monitorDirectOutput(pipe io.Reader) {
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "FATAL") || strings.HasPrefix(line, "ERROR") {
			panic(fmt.Sprintf("Container log error: %s\n", line))
		}
	}
}
