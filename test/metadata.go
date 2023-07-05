package test

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"strings"
	"testing"
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
