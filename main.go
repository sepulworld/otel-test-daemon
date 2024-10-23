package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	datadogstatsd "github.com/DataDog/datadog-go/statsd"
	cactusstatsd "github.com/cactus/go-statsd-client/statsd"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	metrictype "go.opentelemetry.io/otel/metric"
	metricsdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

var (
	datadogReceiver = flag.String("datadog-receiver", "127.0.0.1:8126", "Datadog receiver endpoint")
	httpReceiver    = flag.String("http-receiver", "127.0.0.1:4318", "OpenTelemetry HTTP receiver endpoint")
	lokiReceiver    = flag.String("loki-receiver", "127.0.0.1:3100", "Loki receiver endpoint")
	statsdReceiver  = flag.String("statsd-receiver", "127.0.0.1:9126", "StatsD receiver endpoint")
)

func main() {
	flag.Parse()

	// Run as a daemon
	go func() {
		log.Println("Daemon started")
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Daemon stopping")
		os.Exit(0)
	}()

	// Setup OpenTelemetry Tracer
	if !isPortOpen(*httpReceiver) {
		log.Fatalf("HTTP receiver port is not open: %s", *httpReceiver)
	}
	traceExporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpoint(*httpReceiver), otlptracehttp.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to create trace exporter: %v", err)
	}
	res, err := resource.New(context.Background(), resource.WithAttributes(
		attribute.String("service.name", "otel-test-daemon"),
	))
	if err != nil {
		log.Fatalf("Failed to create resource: %v", err)
	}
	tracerProvider := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(traceExporter),
		tracesdk.WithResource(res),
	)
	otel.SetTracerProvider(tracerProvider)
	tracer := otel.Tracer("otel-test-daemon")

	// Setup OpenTelemetry Metrics (using a basic stdout exporter for now)
	meterProvider := metricsdk.NewMeterProvider()
	otel.SetMeterProvider(meterProvider)
	meter := meterProvider.Meter("otel-test-daemon")

	// Create a counter metric
	counter, err := meter.Float64Counter(
		"test_counter",
		metrictype.WithDescription("A test counter metric"),
	)
	if err != nil {
		log.Fatalf("Failed to create counter metric: %v", err)
	}

	statsdClient, err := cactusstatsd.NewBufferedClient(*statsdReceiver, "otel-test-daemon", 300*time.Millisecond, 0)
	if err != nil {
		log.Fatalf("Failed to create StatsD client: %v", err)
	}
	defer statsdClient.Close()

	// Setup Datadog client
	if !isPortOpen(*datadogReceiver) {
		log.Fatalf("Datadog receiver port is not open: %s", *datadogReceiver)
	}
	datadogClient, err := datadogstatsd.New(*datadogReceiver, datadogstatsd.WithNamespace("otel-test-daemon"))
	if err != nil {
		log.Fatalf("Failed to create Datadog client: %v", err)
	}
	defer datadogClient.Close()

	// Start generating telemetry data
	go func() {
		for {
			// Send a trace
			if err := sendTestTrace(tracer); err != nil {
				log.Printf("Failed to send trace: %v", err)
			}

			// Send a metric
			if err := sendTestMetric(counter); err != nil {
				log.Printf("Failed to send metric: %v", err)
			}

			// Simulate log generation
			if err := sendLokiMessage(*lokiReceiver); err != nil {
				log.Printf("Failed to send Loki message: %v", err)
			}

			// Send StatsD metric
			if err := sendTestStatsdMetric(statsdClient); err != nil {
				log.Printf("Failed to send StatsD metric: %v", err)
			}

			// Send Datadog metric
			if err := sendTestDatadogMetric(datadogClient); err != nil {
				log.Printf("Failed to send Datadog metric: %v", err)
			}

			// Wait for 5 seconds before sending the next batch of test data
			time.Sleep(5 * time.Second)
		}
	}()

	// Block forever
	select {}
}

func sendTestTrace(tracer trace.Tracer) error {
	_, span := tracer.Start(context.Background(), "TestSpan")
	defer span.End()
	span.SetAttributes(
		attribute.String("receiver", *httpReceiver),
		attribute.Float64("test.value", rand.Float64()),
	)
	log.Println("Test trace sent to HTTP receiver", *httpReceiver)
	return nil
}

func sendTestMetric(counter metrictype.Float64Counter) error {
	counter.Add(context.Background(), rand.Float64(),
		metrictype.WithAttributes(attribute.String("endpoint", *httpReceiver)))
	log.Println("Test metric sent to HTTP receiver", *httpReceiver)
	return nil
}

func sendTestStatsdMetric(client cactusstatsd.Statter) error {
	err := client.Gauge("test_gauge", int64(rand.Intn(100)), 1.0)
	if err != nil {
		return err
	}
	log.Println("Test StatsD metric sent to receiver", *statsdReceiver)
	return nil
}

func sendTestDatadogMetric(client *datadogstatsd.Client) error {
	err := client.Gauge("test.datadog.gauge", rand.Float64()*100, []string{"environment:test"}, 1.0)
	if err != nil {
		return err
	}
	log.Println("Test Datadog metric sent to receiver", *datadogReceiver)
	return nil
}

func sendLokiMessage(address string) error {
	lokiURL := fmt.Sprintf("http://%s/loki/api/v1/push", address)

	timeStamp := time.Now().UnixNano()
	logLine := "Test Loki message"

	payload := map[string]interface{}{
		"streams": []map[string]interface{}{
			{
				"stream": map[string]string{
					"app":       "otel-test-daemon",
					"log_level": "info",
				},
				"values": [][]interface{}{
					{fmt.Sprintf("%d", timeStamp), logLine},
				},
			},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", lokiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		log.Printf("Loki push successful, response status: %s", resp.Status)
		return nil
	}

	log.Println("Test Loki message sent to receiver", address)
	return nil
}

func isPortOpen(address string) bool {
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
