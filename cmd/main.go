package main

import (
	"hexagonal/internal/adapters/event"
	"hexagonal/internal/adapters/handler"
	"hexagonal/internal/adapters/repository"
	"hexagonal/internal/core/services"
	"hexagonal/pkg/config"
	"hexagonal/pkg/logger"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log := logger.NewLogger()
	defer log.Sync()

	zap.L().Info("starting application...",
		zap.String("env", cfg.AppEnv),
		zap.String("port", cfg.ServerPort),
	)

	repo := repository.NewInMemoryRepository()
	// Config'den gelen broker listesini ve topic adını veriyoruz.
	publisher, err := event.NewKafkaPublisher(cfg.GetBrokers(), cfg.KafkaTopic)
	if err != nil {
		// Kafka bağlantısı kritik olduğu için bağlanamazsa uygulamayı kapatıyoruz (Fail Fast).
		zap.L().Fatal("failed to connect to kafka", zap.Error(err))
	}
	defer publisher.Close()
	svc := services.NewConcertService(repo, publisher)
	myHandler := handler.NewHTTPHandler(svc)

	app := fiber.New()
	app.Post("/concerts", myHandler.CreateConcert)
	app.Post("/tickets", myHandler.BuyTicket)
	zap.L().Info("server is listening", zap.String("port", cfg.ServerPort))

	if err := app.Listen(":" + cfg.ServerPort); err != nil {
		zap.L().Fatal("server failed to start", zap.Error(err))
	}
}
