# syntax=docker/dockerfile:1

# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o go-grafana main.go

# Final image
FROM gcr.io/distroless/base
WORKDIR /app
COPY --from=builder /app/go-grafana ./go-grafana
# config/.env should be mounted at runtime
EXPOSE 2112
CMD ["./go-grafana"] 