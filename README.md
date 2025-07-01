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
```

- `URLS`: Comma-separated list of URLs to monitor
- `CHECK_INTERVAL`: How often to check the URLs (e.g., `60s`, `5m`). Default is 51s if unset.
- `SMTP_SERVER`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`: SMTP server details for sending email
- `SMTP_TO`: Recipient email address
- `SMTP_FROM`: Sender email address

## Features

- Monitors HTTP status of configured URLs
- Exposes Prometheus metrics at `/metrics`
- Sends email alerts when a site goes offline or recovers
- Logs alert and recovery events
- Graceful shutdown on SIGINT/SIGTERM
