package test

import (
	"testing"
)

func AppWorks(t *testing.T) {
	t.Run("TestAppInitialization", AppIsInitialized)
	t.Run("TestAppStore", CanAddAppToStore)
	t.Run("TestIconFile", IconFileIsValid)
	t.Run("TestVersionEndpoint", VersionEndpointExists)
	t.Run("TestAPISpecEndpoint", APISpecEndpointExists)
}
