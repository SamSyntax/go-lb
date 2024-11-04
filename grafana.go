package main

import (
	"context"

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"

	// "go.opentelemetry.io/otel/expoters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer       trace.Tracer
	otlpEndpoint string
)

func epInit() {
	otlpEndpoint = "localhost:4318"
	if otlpEndpoint == "" {
		log.Fatalln("Provide OTLP_ENDPOINT env variable")
	}
}

// func newConsoleExporter() (oteltrace.SpanExporter, error) {
// 	return stdouttrace.New()
// }

func newOTLPExporter(ctx context.Context) (oteltrace.SpanExporter, error) {
	insecureOpt := otlptracehttp.WithInsecure()

	endpointOpt := otlptracehttp.WithEndpoint(otlpEndpoint)
	return otlptracehttp.New(ctx, insecureOpt, endpointOpt)
}

func newTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	r, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("go-lb"),
		),
	)
	if err != nil {
		log.Errorf("NewTraceProvider: %v", err)
		panic(err)
	}
	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(r),
	)
}
