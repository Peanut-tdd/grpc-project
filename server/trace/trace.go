package trace

import (
	"context"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	newTrace "go.opentelemetry.io/otel/trace"
)

const (
	jaegerEndpoint = "127.0.0.1:4317"
	Service        = "grpc-server"
	Env            = "dev"
)

func InitTracer(ctx context.Context) func() {
	url := os.Getenv("JAEGER_ENDPOINT")
	if url == "" {
		url = jaegerEndpoint
	}

	//exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))

	// 创建一个使用 HTTP 协议连接本机Jaeger的 Exporter
	exp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(url),
		otlptracegrpc.WithInsecure())

	if err != nil {
		return nil
	}

	tp := tracesdk.NewTracerProvider(
		tracesdk.WithSampler(tracesdk.AlwaysSample()),
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(Service),
			attribute.String("environment", Env),
		),
		),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return func() { _ = tp.Shutdown(context.Background()) }
}

func FuncCall(ctx context.Context, spanName string) context.Context {

	tracer := otel.Tracer(Service)
	ctx, span := tracer.Start(ctx, spanName,
		newTrace.WithAttributes(attribute.String("app", Service)))
	defer span.End()

	return ctx
}
