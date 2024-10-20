.PHONY: all build run clean

all: build

# Build the Go binary
build:
	@echo "Building the Go application..."
	GOOS=linux GOARCH=amd64 go build -o otel-test-daemon

# Run the application locally
run:
	@echo "Running the Go application..."
	./otel-test-daemon

# Build the Docker image
image:
	@echo "Building Docker image..."
	docker build -t otel-test-daemon:latest .

# Clean up generated files
clean:
	@echo "Cleaning up..."
	rm -f otel-test-daemon
