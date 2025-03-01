package amqp

import (
	"cicd-processor/domain"
	"cicd-processor/lib"
	"cicd-processor/lib/logger/sl"
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"log/slog"

	"github.com/rabbitmq/amqp091-go"
)

// Consumer отвечает за подключение к RabbitMQ и отправку сообщений в канал
type Consumer struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
	queue   string
	log     *slog.Logger
}

// NewConsumer создает нового потребителя
func NewConsumer(amqpURL, queueName string, log *slog.Logger) (*Consumer, error) {
	log.Info("Create Consumer")
	conn, err := amqp091.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	_, err = ch.QueueDeclare(
		queueName, true, false, false, false, nil,
	)
	if err != nil {
		conn.Close()
		ch.Close()
		return nil, err
	}

	return &Consumer{
		conn:    conn,
		channel: ch,
		queue:   queueName,
		log:     log,
	}, nil
}

// StartListening запускает прослушивание очереди и отправку сообщений в канал
func (c *Consumer) StartListening(messageChannel chan domain.UpdateMessageWithContext) {
	const op = "Consumer.StartListening"
	log := c.log.With(
		slog.String("op", op),
	)
	log.Info("Start listening")
	msgs, err := c.channel.Consume(
		c.queue, "", false, false, false, false, nil,
	)
	if err != nil {
		log.Error("Ошибка подписки на очередь:", sl.Err(err))
	}

	// Обрабатываем каждое сообщение в горутине
	go func() {
		for msg := range msgs {

			ctx, span := tracerStart(msg.Headers)

			uuid := lib.GetUUIDFromHeaders(msg.Headers)
			span.SetAttributes(attribute.String("uuid", uuid))

			update, err := domain.ParseMessage(msg.Body)
			if err != nil {
				log.Error("Ошибка декодирования JSON: %s", sl.Err(err))
				_ = msg.Ack(false)
				continue
			}
			updateWithContext := domain.UpdateMessageWithContext{
				UpdateMessage: update,
				UUID:          uuid,
				Context:       ctx,
			}

			// Пишем сообщение в канал
			messageChannel <- updateWithContext
			_ = msg.Ack(true) // Подтверждаем получение
			span.End()
		}
	}()
}

func (c *Consumer) IsQueueOk() (string, error) {
	_, err := c.channel.QueueDeclarePassive(
		c.queue, true, false, false, false, nil,
	)
	if err != nil {
		return "DOWN", err
	}
	return "UP", nil
}

func tracerStart(msgHeaders amqp091.Table) (context.Context, trace.Span) {

	// Конвертируем amqp091.Table в propagation.MapCarrier
	carrier := lib.MapAMQPTableToMapCarrier(msgHeaders)

	// Восстанавливаем контекст OpenTelemetry
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)

	tracer := otel.Tracer("tg-sender")
	return tracer.Start(ctx, "MessageReceived")
}

// Close закрывает соединение с RabbitMQ
func (c *Consumer) Close() {
	const op = "Consumer.Close"

	c.log.Info("Close consumer", slog.String("op", op))
	c.channel.Close()
	c.conn.Close()
}
