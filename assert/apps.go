package assert

import (
	"testing"
)

func AppWorks(t *testing.T) bool {
	IconFileIsValid(t)
	VersionEndpointExists(t)
	return t.Failed()
}
