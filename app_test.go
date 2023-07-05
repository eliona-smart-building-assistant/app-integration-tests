package main

import (
	"github.com/eliona-smart-building-assistant/app-integration-tests/app"
	"github.com/eliona-smart-building-assistant/app-integration-tests/test"
	"testing"
)

func TestMain(m *testing.M) {
	app.RunApp(m)
}

func TestAppWorks(t *testing.T) {
	test.AppWorks(t)
}
