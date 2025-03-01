package lib

import (
	"context"
	"github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

func GetUUIDFromHeaders(msgHeaders amqp091.Table) string {
	if value, ok := msgHeaders["uuid"].(string); ok {
		return value
	}
	const ZeroUUID = "00000000-0000-4000-8000-000000000000"
	return ZeroUUID
}

func MapAMQPTableToMapCarrier(headers amqp091.Table) propagation.MapCarrier {
	carrier := propagation.MapCarrier{}
	for key, value := range headers {
		if str, ok := value.(string); ok { // Проверяем, является ли значение строкой
			carrier[key] = str
		}
	}
	return carrier
}

func MapCarrierToAMQPTable(ctx context.Context) amqp091.Table {
	headers := amqp091.Table{}
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	for key, value := range carrier {
		headers[key] = value
	}
	return headers
}
