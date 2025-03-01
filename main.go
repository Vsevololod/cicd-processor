package main

import (
	"cicd-processor/communication/amqp"
	"cicd-processor/config"
	"cicd-processor/domain"
	"cicd-processor/lib/logger/sl"
	"cicd-processor/service"
	"cicd-processor/tracing"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {

	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)
	log.Info("Starting Telegram Sender")

	shutdownTracer := tracing.InitTracer(&cfg.OtlpConfig)
	defer shutdownTracer()

	inputMessageChannel := make(chan domain.UpdateMessageWithContext, 100)

	consumer := registerConsumer(inputMessageChannel, &cfg.AmqpConf, log)

	processService := service.NewMessageProcessService(inputMessageChannel, cfg.PathsConf, log)
	processService.StartProcessing(5)

	health := NewHealth(consumer, log)
	health.Start()

	// Контекст для graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done() // Ждем сигнала завершения

	log.Info("Shutdown signal received. Closing services...")

	// Закрываем consumer
	consumer.Close()
	log.Info("Consumer stopped.")

	// Закрываем процесс обработки сообщений
	processService.StopProcessing()
	log.Info("Message processing stopped.")

	health.StopProcessing()
	log.Info("Stop /health endpoint.")

	log.Info("Shutdown complete.")

}

func registerConsumer(inputMessageChannel chan domain.UpdateMessageWithContext, cfg *config.AmqpConfig, log *slog.Logger) *amqp.Consumer {
	consumer, err := amqp.NewConsumer(cfg.GetAmqpUri(), cfg.QueueName, log)
	if err != nil {
		log.Error("Ошибка создания потребителя:", sl.Err(err))
		os.Exit(1)
	}

	go consumer.StartListening(inputMessageChannel)
	return consumer
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
