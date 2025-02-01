# Use the official Golang image as the base image for the build stage
FROM golang:1.23 as builder

COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o otel-test-daemon

# Use the same Golang base image for the runtime to ensure compatibility
FROM golang:1.23

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /go/otel-test-daemon .

# Define default values for command-line arguments
ARG DATADOG_RECEIVER="169.254.1.1:8126"
ARG HTTP_RECEIVER="169.254.1.1:4318"
ARG SYSLOG_RECEIVER="169.254.1.1:51893"
ARG STATSD_RECEIVER="169.254.1.1:9126"

# Set environment variables from the arguments
ENV DATADOG_RECEIVER=${DATADOG_RECEIVER}
ENV HTTP_RECEIVER=${HTTP_RECEIVER}
ENV SYSLOG_RECEIVER=${SYSLOG_RECEIVER}
ENV STATSD_RECEIVER=${STATSD_RECEIVER}

# Command to run the executable with environment variables as arguments
CMD ./otel-test-daemon -datadog-receiver=${DATADOG_RECEIVER} -http-receiver=${HTTP_RECEIVER} -tcp-syslog=${SYSLOG_RECEIVER} -statsd-receiver=${STATSD_RECEIVER}
