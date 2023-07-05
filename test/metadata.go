package test

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"github.com/eliona-smart-building-assistant/app-integration-tests/app"
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
	metadata, metadataData, err := app.GetMetadata()
	require.NoError(t, err, "Getting metadata successful")

	db, err := app.GetDB()
	require.NoError(t, err, "Connect to database")

	iconFile, err := os.Open("icon")
	require.NoError(t, err, "Opening icon file")
	defer iconFile.Close()

	iconData, err := io.ReadAll(iconFile)
	require.NoError(t, err, "Reading icon file")

	if _, err := db.Exec(`
		UPDATE eliona_store
		SET metadata = $1, icon = $2
		WHERE app_name = $3`, string(metadataData), string(iconData), metadata.Name); err != nil {
		assert.NoError(t, err, "executing SQL script: %s")
	}
}

func AppIsInitialized(t *testing.T) {
	metadata, _, err := app.GetMetadata()
	require.NoError(t, err, "Getting metadata successful")

	db, err := app.GetDB()
	require.NoError(t, err, "Connect to database")

	row := db.QueryRow(`
		SELECT initialized_at
		FROM public.eliona_app
		WHERE app_name = $1;
	`, metadata.Name)
	var initialized *time.Time
	err = row.Scan(&initialized)
	require.NoError(t, err, "executing select statement")
	assert.NotEmpty(t, initialized, "initialized_at shouldn't be empty")
}
