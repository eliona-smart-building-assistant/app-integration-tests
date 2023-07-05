package main

import (
	"github.com/eliona-smart-building-assistant/app-integration-tests/docker"
	"github.com/eliona-smart-building-assistant/app-integration-tests/test"
	"testing"
)

func TestMain(m *testing.M) {
	docker.RunApp(m)
}

func TestApp(t *testing.T) {
	test.AppWorks(t)
}
