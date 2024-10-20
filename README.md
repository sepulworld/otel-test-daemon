# otel-test-daemon

## Overview

`otel-test-daemon` is a Go application that generates and sends test metrics, traces, and logs to various telemetry systems including Datadog, OpenTelemetry, Syslog, and StatsD. This project aims to simulate telemetry data for testing observability and monitoring setups.

The application runs as a daemon, periodically generating telemetry data to configured endpoints. It uses OpenTelemetry for metrics and traces, and integrates with Datadog, StatsD, and Syslog.

## Features

- Sends telemetry data to multiple systems:
  - OpenTelemetry receiver (HTTP endpoint)
  - Datadog (via StatsD client)
  - Syslog
  - StatsD receiver
- Configurable through command-line flags.
- Runs as a daemon, automatically generating telemetry data.
- Dockerized for easy deployment.

## Requirements

- Go 1.19+
- Docker (for containerized deployments)
- GitHub account (for using GitHub Actions to build and release Docker images)

## Installation

### Cloning the Repository

```sh
git clone https://github.com/sepulworld/otel-test-daemon.git
cd otel-test-daemon
```

### Building Locally

You can build the application using Go:

```sh
go build -o otel-test-daemon
```

### Running the Application

```sh
./otel-test-daemon
```

You can configure the receivers using command-line flags:

```sh
./otel-test-daemon \
  -datadog-receiver=169.254.1.1:8126 \
  -http-receiver=169.254.1.1:4318 \
  -syslog-receiver=169.254.1.1:51893 \
  -statsd-receiver=169.254.1.1:9126
```

## Docker

### Building the Docker Image

You can build the Docker image for `otel-test-daemon` using the following command:

```sh
docker build -t ghcr.io/sepulworld/otel-test-daemon:latest -f Dockerfile .
```

### Running the Docker Container

Run the container with:

```sh
docker run -d ghcr.io/sepulworld/otel-test-daemon:latest
```

## GitHub Actions

This project includes a GitHub Actions workflow to automatically build and release Docker images to the GitHub Container Registry (GHCR) whenever there is a merge to the `main` branch or a new Git tag is created.

## Configuration

The application can be configured via the following command-line flags:

- `-datadog-receiver` (default: `169.254.1.1:8126`): Datadog receiver endpoint.
- `-http-receiver` (default: `169.254.1.1:4318`): OpenTelemetry HTTP receiver endpoint.
- `-syslog-receiver` (default: `169.254.1.1:51893`): Syslog receiver endpoint.
- `-statsd-receiver` (default: `169.254.1.1:9126`): StatsD receiver endpoint.

## Development

### Prerequisites

- Go 1.19+
- Docker (optional, for containerized deployment)

### Running Tests

Unit tests can be run using the Go test tool:

```sh
go test ./...
```

## GitHub Workflow

The GitHub Actions workflow file (`.github/workflows/build-and-release.yml`) is set up to:

- Build the Docker image.
- Push the Docker image to the GitHub Container Registry (GHCR).
- Create a new release when a Git tag is created.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request for any improvements.

## Contact

For any questions or suggestions, feel free to reach out to the project maintainer.
