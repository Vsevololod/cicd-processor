package tracing

import (
	"cicd-processor/config"
	"context"
	"fmt"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.21.0"
)

// InitTracer настраивает OpenTelemetry с OTLP gRPC Exporter
func InitTracer(cfg *config.OtlpConfig) func() {
	ctx := context.Background()

	// Создаем OTLP Trace экспортер через gRPC
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(), // Без TLS (для тестов)
		otlptracegrpc.WithEndpoint(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)),
	)
	if err != nil {
		log.Fatalf("Ошибка создания OTLP Trace Exporter: %v", err)
	}

	// Создаем TracerProvider
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()), // Логируем все трейсы
		trace.WithBatcher(exporter),             // Отправляем трейсы батчами
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName), // Имя сервиса
		)),
	)

	otel.SetTracerProvider(tp) // Устанавливаем глобальный провайдер

	return func() {
		_ = tp.Shutdown(ctx)
	}
}
