package main

import (
	"github.com/eliona-smart-building-assistant/app-integration-tests/assert"
	"github.com/eliona-smart-building-assistant/app-integration-tests/docker"
	"testing"
)

func TestMain(m *testing.M) {
	docker.RunApp(m)
}

func TestApp(t *testing.T) {
	assert.AppWorks(t)
}
