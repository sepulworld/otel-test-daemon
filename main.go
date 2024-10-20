package main

import (
	"context"
	"flag"
	"log"
	"math/rand"
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
	datadogReceiver = flag.String("datadog-receiver", "169.254.1.1:8126", "Datadog receiver endpoint")
	httpReceiver    = flag.String("http-receiver", "169.254.1.1:4318", "OpenTelemetry HTTP receiver endpoint")
	syslogReceiver  = flag.String("syslog-receiver", "169.254.1.1:51893", "Syslog receiver endpoint")
	statsdReceiver  = flag.String("statsd-receiver", "169.254.1.1:9126", "StatsD receiver endpoint")
)

func main() {
	// Run as a daemon
	go func() {
		log.Println("Daemon started")
		// Capture system interrupts to gracefully shutdown
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		log.Println("Daemon stopping")
		os.Exit(0)
	}()

	// Setup OpenTelemetry Tracer
	traceExporter, err := otlptracehttp.New(context.Background(), otlptracehttp.WithEndpoint(*httpReceiver))
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

	// Setup StatsD client
	statsdClient, err := cactusstatsd.NewBufferedClient(*statsdReceiver, "otel-test-daemon", 300*time.Millisecond, 0)
	if err != nil {
		log.Fatalf("Failed to create StatsD client: %v", err)
	}
	defer statsdClient.Close()

	// Setup Datadog client
	datadogClient, err := datadogstatsd.New(*datadogReceiver, datadogstatsd.WithNamespace("otel-test-daemon"))
	if err != nil {
		log.Fatalf("Failed to create Datadog client: %v", err)
	}
	defer datadogClient.Close()

	// Start generating telemetry data
	go func() {
		for {
			// Send a trace
			sendTestTrace(tracer)

			// Send a metric
			counter.Add(context.Background(), rand.Float64(),
				metrictype.WithAttributes(attribute.String("endpoint", *httpReceiver)))

			// Simulate log generation
			log.Println("Sending test log to syslog receiver", *syslogReceiver)

			// Send StatsD metric
			sendTestStatsdMetric(statsdClient)

			// Send Datadog metric
			sendTestDatadogMetric(datadogClient)

			// Wait for 5 seconds before sending the next batch of test data
			time.Sleep(5 * time.Second)
		}
	}()

	// Block forever
	select {}
}

func sendTestTrace(tracer trace.Tracer) {
	_, span := tracer.Start(context.Background(), "TestSpan")
	defer span.End()
	span.SetAttributes(
		attribute.String("receiver", *httpReceiver),
		attribute.Float64("test.value", rand.Float64()),
	)
	log.Println("Test trace sent to HTTP receiver", *httpReceiver)
}

func sendTestStatsdMetric(client cactusstatsd.Statter) {
	err := client.Gauge("test_gauge", int64(rand.Intn(100)), 1.0)
	if err != nil {
		log.Printf("Failed to send StatsD metric: %v", err)
	} else {
		log.Println("Test StatsD metric sent to receiver", *statsdReceiver)
	}
}

func sendTestDatadogMetric(client *datadogstatsd.Client) {
	err := client.Gauge("test.datadog.gauge", rand.Float64()*100, []string{"environment:test"}, 1.0)
	if err != nil {
		log.Printf("Failed to send Datadog metric: %v", err)
	} else {
		log.Println("Test Datadog metric sent to receiver", *datadogReceiver)
	}
}
