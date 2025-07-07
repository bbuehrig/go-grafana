# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]
- Add example `.env` file for easier configuration.
- Use github.com/jordan-wright/email for robust SMTP with STARTTLS support (fixes EOF errors with modern SMTP servers, improves email reliability).

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