# Eliona App Test Suite

This test suite is designed to validate the functionality and behavior of Go applications for Eliona. It checks the main features of the app, and verifies the correctness of database interactions, configuration, metadata, icons, and endpoints.

This test suite currently tests the aspects of apps that are common to all apps. Its main purpose is to simplify and standardize the review of apps before releasing.

## Getting Started

### Environment Variables

The test suite requires the following environment variables:

- `API_ENDPOINT`: The Eliona API endpoint.
- `API_TOKEN`: The Eliona API token.
- `CONNECTION_STRING`: The connection string for the PostgreSQL database.

### How to Run

You can run the tests using the `go test` command. You need to specify the `-app` flag, which is the path to the root directory of the tested app, and optionally the `-test.v` flag if you want verbose output.

```shell
go test -app=/path/to/app -test.v
```

This command will build a Docker image from your Dockerfile, run the container, and then run the test suite against it.

## Directory Structure

- `main_test.go`: This is the main test file. It contains the setup, tear-down, and the `TestMain` function which orchestrates the testing process. The Docker image is built and run, and the environment is checked and initialized in this file.
- `metadata_test.go`: This file contains tests for the metadata. It verifies the presence and correctness of required files.
- `endpoints_test.go`: The name is self-explanatory.

## Handling Failures

The tests could fail already during the initialization process, but should always tell the reason. After the Docker container starts, its output is watched, and any error reported fails the tests.

Remember that the teardown process will always stop the Docker container, even if a test fails or an error occurs during the testing process. It is important to ensure the container is stopped after the tests are run to free up Docker namespace.

## Future development

The current goal of this suite is to take over most of the app review checklist. In the future, to allow test-driven development, this test suite

Happy testing!

