# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]
### Added
- Example `.env` file for easier configuration.
- Email subject now includes the website URL, error code/reason, and a status emoji (🚨 for down, ✅ for up) for clearer alerts and recoveries.
- Email sending logic is now abstracted via an `EmailSender` interface for testability, with unit tests for alert subject/body logic using a mock sender.
- Comprehensive unit tests for configuration loading, metrics initialization, monitoring and alerting logic, and service initialization.
- Configurable alert threshold: only send a DOWN alert after N consecutive failures (set via `ALERT_THRESHOLD`, default 2).
- Basic integration test for the Dagger pipeline in `ci/main_test.go`. The pipeline logic was refactored into a testable function (`RunPipeline`).
### Changed
- Use github.com/jordan-wright/email for robust SMTP with STARTTLS support (fixes EOF errors with modern SMTP servers, improves email reliability).

## [0.3.0] - 2024-06-10
### Fixed
- Prevent duplicate email notifications for site down/recovery events by synchronizing access to offlineMap with a mutex. Now, only one email is sent per state change, even with parallel checks.

## [0.2.0] - 2024-06-10
### Added
- Dagger pipeline (`ci/main.go`) for building and exporting Docker images locally for both amd64 and arm64 architectures.
- Multi-arch image support in the CI pipeline.
- Expanded README with Docker and Dagger usage instructions.
### Changed
- Updated `Dockerfile` to use Go 1.24 for compatibility with `go.mod` requirements.

## [0.1.0] - 2024-06-09
### Added
- Initial release: Go-Grafana Webpage Monitor with Prometheus metrics and email alerting.
- Dockerfile for containerized builds.
- Basic documentation in README. 