# Use the official Golang image as the base image for the build stage
FROM golang:1.23 as builder

COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o otel-test-daemon

# Use a minimal Docker image to package the built binary
FROM debian:bullseye-slim

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/otel-test-daemon .

# Command to run the executable
CMD ["./otel-test-daemon"]

# Expose necessary ports
EXPOSE 4318 8126 51893 9126
