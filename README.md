# Go-Grafana Webpage Monitor

This service monitors a list of webpages and exposes Prometheus metrics. It now also sends email alerts when a site goes offline or recovers.

## Configuration

Create a `config/.env` file with the following variables:

```
URLS=https://example.com,https://another.com
CHECK_INTERVAL=60s
SMTP_SERVER=smtp.example.com
SMTP_PORT=587
SMTP_USER=youruser@example.com
SMTP_PASS=yourpassword
SMTP_TO=alertrecipient@example.com
SMTP_FROM=monitor@example.com
ALERT_THRESHOLD=2
```

- `URLS`: Comma-separated list of URLs to monitor
- `CHECK_INTERVAL`: How often to check the URLs (e.g., `60s`, `5m`). Default is 51s if unset.
- `SMTP_SERVER`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`: SMTP server details for sending email
- `SMTP_TO`: Recipient email address
- `SMTP_FROM`: Sender email address
- `ALERT_THRESHOLD`: Number of consecutive failures before sending a DOWN alert (default: 2)

## Features

- Monitors HTTP status of configured URLs
- Exposes Prometheus metrics at `/metrics`
- Sends email alerts when a site goes offline or recovers
- Email alert subject includes the website URL, error code/reason, and a status emoji (🚨 for down, ✅ for up)
- Logs alert and recovery events
- Graceful shutdown on SIGINT/SIGTERM
- Configurable alert threshold: only sends a DOWN alert after N consecutive failures (set via `ALERT_THRESHOLD`, default 2)

### Email Alert Subject Format

When a site goes down, the email subject will look like:

```
[🚨 DOWN] https://example.com (returned status 500)
```

When a site recovers, the subject will look like:

```
[✅ UP] https://example.com is back online
```

## Docker Usage

A multi-stage `Dockerfile` is provided for building and running the service in a containerized environment.

### Build and Run with Docker

1. Build the Docker image:
   ```sh
   docker build -t go-grafana-monitor:latest .
   ```
2. Run the container (mount your config as needed):
   ```sh
   docker run -d -p 2112:2112 -v $(pwd)/config/.env:/app/config/.env go-grafana-monitor:latest
   ```

- The service will be available at `http://localhost:2112/metrics`.
- Make sure to provide the required `config/.env` file (see Configuration section).

## Dagger Pipeline (CI)

This project includes a Dagger-based pipeline for building and exporting Docker images for multiple architectures (amd64 and arm64) locally.

- The pipeline is defined in [`ci/main.go`](ci/main.go).
- It uses the [dagger.io/dagger](https://dagger.io/) Go SDK.
- **Note:** The pipeline is pinned to Dagger v0.17.2 for stability due to OpenTelemetry compatibility issues with later versions.

### Running the Dagger Pipeline

1. Install Dagger Go SDK:
   ```sh
   go get dagger.io/dagger@v0.17.2
   ```
2. Run the pipeline:
   ```sh
   go run ci/main.go
   ```

This will build the Docker image for both `amd64` and `arm64` architectures using the local `Dockerfile` and export them locally as:
- `go-grafana-monitor:amd64`
- `go-grafana-monitor:arm64`

You can then load and run the images with Docker (on the appropriate architecture), for example:
```sh
docker run -d -p 2112:2112 go-grafana-monitor:amd64
# or
docker run -d -p 2112:2112 go-grafana-monitor:arm64
```

### Testing the Dagger Pipeline

A basic integration test for the Dagger pipeline is provided in [`ci/main_test.go`](ci/main_test.go`). This test attempts to run the pipeline using a real Dagger engine. **By default, the test is skipped.**

To run the pipeline test, set the environment variable `CI_DAGGER_TEST=1`:

```sh
CI_DAGGER_TEST=1 go test ./ci
```

- The test requires a running Dagger engine and Docker available on your system.
- The test will fail if the pipeline fails to build or export the images.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for a detailed list of recent changes and improvements.

## Testability

The project includes comprehensive unit tests for configuration loading, metrics initialization, monitoring and alerting logic, and service initialization, in addition to email subject and sending logic. Tests are located in:
- `email_test.go`
- `metrics_test.go`
- `config_test.go`
- `monitor_test.go`
- `main_test.go`

The email sending logic is abstracted via an `EmailSender` interface, allowing for unit testing of alert and recovery logic without sending real emails. See the test files for examples of how the alert subject, body, and monitoring transitions are verified using mocks and real Prometheus metrics.
