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
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/eliona-smart-building-assistant/go-eliona/app"
	"github.com/eliona-smart-building-assistant/go-utils/db"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	appLocation string
)

func RunApp(m *testing.M) {
	StartApp()
	m.Run()
	StopApp()
}

func StartApp() {
	handleEnvironment()
	resetDB()
	switch StartMode() {
	case StartModeDirect:
		startAppDirectly()
	case StartModeDocker:
		startAppContainer()
	}
}

func StopApp() {
	switch StartMode() {
	case StartModeDirect:
		stopAppDirectly()
	case StartModeDocker:
		stopAppContainer()
	}
}

const (
	StartModeDirect string = "direct"
	StartModeDocker string = "docker"
)

func StartMode() string {
	mode, present := os.LookupEnv("START_MODE")
	if present {
		return strings.ToLower(mode)
	}
	return StartModeDocker
}

func handleFlags() {
	flag.StringVar(&appLocation, "app", "", "Path to app")
	flag.Parse()

	if appLocation == "" {
		appLocation = "."
	}
}

func handleEnvironment() {
	if err := checkEnvVars(); err != nil {
		fmt.Printf("checking environment variables: %v", err)
		os.Exit(1)
	}

	handleFlags()
	if err := os.Chdir(appLocation); err != nil {
		fmt.Printf("chdir to %s: %v", appLocation, err)
		os.Exit(1)
	}
}

func checkEnvVars() error {
	var present bool

	_, present = os.LookupEnv("API_ENDPOINT")
	if !present {
		return errors.New("API_ENDPOINT variable not defined")
	}

	_, present = os.LookupEnv("API_TOKEN")
	if !present {
		return errors.New("API_TOKEN variable not defined")
	}

	_, present = os.LookupEnv("CONNECTION_STRING")
	if !present {
		return errors.New("CONNECTION_STRING variable not defined")
	}
	return nil
}

func waitForAppReady() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	fmt.Println("Waiting for the app to get ready...")

	metadata, _, err := app.GetMetadata()
	if err != nil {
		return fmt.Errorf("getting metadata: %s", err)
	}

	url := fmt.Sprintf("http://localhost:3039/%s/version", metadata.ApiUrl)
	client := &http.Client{Timeout: 100 * time.Millisecond}
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("app did not become ready in the specified time")
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

func resetDB() {
	metadata, _, err := app.GetMetadata()
	if err != nil {
		fmt.Printf("getting metadata: %s", err)
		os.Exit(1)
	}

	database := db.NewDatabase("dd")

	sqlScript, err := os.ReadFile("reset.sql")
	if err != nil {
		fmt.Printf("reading SQL script: %s", err)
		os.Exit(1)
	}

	_, err = database.Exec(string(sqlScript))
	if err != nil {
		fmt.Printf("executing SQL script: %s", err)
		os.Exit(1)
	}

	row := database.QueryRow(`
		SELECT initialized_at
		FROM public.eliona_app
		WHERE app_name = $1;
	`, metadata.Name)
	var initialized *time.Time
	if err = row.Scan(&initialized); err != nil {
		fmt.Printf("executing SELECT statement: %s", err)
		os.Exit(1)
	}

	// Check if the script reset the initialization state of app
	if initialized != nil {
		fmt.Printf("unexpected result from SELECT statement: got %v, want nil", initialized)
		os.Exit(1)
	}
}
