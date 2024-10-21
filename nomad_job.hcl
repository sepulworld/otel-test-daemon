job "otel-test-daemon" {
  datacenters = ["dc1"]
  type = "service"

  group "otel-test-daemon" {
    count = 1

    task "otel-test-daemon" {
      driver = "docker"

      config {
        image = "ghcr.io/sepulworld/otel-test-daemon:latest"
        args = [
          "-datadog-receiver=169.254.1.1:8126",
          "-http-receiver=169.254.1.1:4318",
          "-syslog-receiver=169.254.1.1:51893",
          "-statsd-receiver=169.254.1.1:9126"
        ]
      }

      resources {
        cpu    = 50
        memory = 128
      }

      service {
        name = "otel-test-daemon"
        port = "http"

        check {
          type     = "tcp"
          port     = "http"
          interval = "10s"
          timeout  = "2s"
        }
      }
    }
  }
}
