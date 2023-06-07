package assert

import (
	"testing"

	"github.com/eliona-smart-building-assistant/app-integration-tests/app"
)

func AppWorks(t *testing.T) bool {
	metadata, err := app.GetMetadata()
	if err != nil {
		t.Fatalf("getting metadata: %v", err)
	}

	IconFileIsValid(t)
	VersionEndpointExists(t, metadata)
	APISpecEndpointExists(t, metadata)

	return t.Failed()
}
