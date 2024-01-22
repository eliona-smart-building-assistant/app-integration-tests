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
	"bufio"
	"bytes"
	"encoding/base64"
	eapp "github.com/eliona-smart-building-assistant/go-eliona/app"
	"github.com/eliona-smart-building-assistant/go-utils/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func IconFileIsValid(t *testing.T) {
	t.Parallel()

	file, err := os.Open("icon")
	if err != nil {
		t.Fatalf("Failed to open icon file: %s", err)
	}
	defer file.Close()

	iconData, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read icon file: %s", err)
	}

	if !strings.HasPrefix(string(iconData), "data:image/png;base64,") && !strings.HasPrefix(string(iconData), "data:image/jpeg;base64,") {
		t.Fatalf("Invalid icon data prefix")
	}

	if len(iconData) > 63*1024 {
		// The size limit for "text" type in db is 64 kB. Let's set the limit a little lower for safety.
		t.Fatalf("Image size is larger than size limit: %d bytes", len(iconData))
	}

	decodedData, err := base64.StdEncoding.DecodeString(strings.SplitN(string(iconData), ",", 2)[1])
	if err != nil {
		t.Fatalf("Failed to decode base64 data: %s", err)
	}

	_, _, err = image.Decode(bufio.NewReader(bytes.NewReader(decodedData)))
	if err != nil {
		t.Fatalf("Failed to decode image data: %s", err)
	}
}

func CanAddAppToStore(t *testing.T) {
	t.Parallel()

	metadata, metadataData, err := eapp.GetMetadata()
	require.NoError(t, err, "Getting metadata successful")

	database := db.NewDatabase("app-integration-test")

	iconFile, err := os.Open("icon")
	require.NoError(t, err, "Opening icon file")
	defer iconFile.Close()

	iconData, err := io.ReadAll(iconFile)
	require.NoError(t, err, "Reading icon file")

	if _, err := database.Exec(`
		UPDATE eliona_store
		SET metadata = $1, icon = $2
		WHERE app_name = $3`, string(metadataData), string(iconData), metadata.Name); err != nil {
		assert.NoError(t, err, "executing SQL script: %s")
	}
}

func AppIsInitialized(t *testing.T) {
	t.Parallel()

	metadata, _, err := eapp.GetMetadata()
	require.NoError(t, err, "Getting metadata successful")

	database := db.NewDatabase("app-integration-test")

	row := database.QueryRow(`
		SELECT initialized_at
		FROM public.eliona_app
		WHERE app_name = $1;
	`, metadata.Name)
	var initialized *time.Time
	err = row.Scan(&initialized)
	require.NoError(t, err, "executing select statement")
	assert.NotEmpty(t, initialized, "initialized_at shouldn't be empty")
}
