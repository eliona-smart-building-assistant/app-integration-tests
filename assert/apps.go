package assert

import (
	"testing"
)

func AppWorks(t *testing.T) {
	t.Run("TestIconFile", IconFileIsValid)
	t.Run("TestVersionEndpoint", VersionEndpointExists)
	t.Run("TestAPISpecEndpoint", APISpecEndpointExists)
}
