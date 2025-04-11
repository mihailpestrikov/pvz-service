package main

import (
	"avito-backend-trainee-assignment-spring-2025/internal/repository/postgres"
	"avito-backend-trainee-assignment-spring-2025/pkg/config"
	"avito-backend-trainee-assignment-spring-2025/pkg/logger"
	"avito-backend-trainee-assignment-spring-2025/pkg/metrics"
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger.Setup(cfg.Logger)

	log.Info().
		Str("app_env", cfg.Server.AppEnv).
		Str("port", cfg.Server.Port).
		Msg("Starting application")

	db, err := postgres.New(&cfg.Postgres)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to database")
	}
	defer db.Close()

	metricsServer := metrics.NewServer(cfg.Prometheus.Port)
	go func() {
		if err := metricsServer.Start(); err != nil {
			log.Error().Err(err).Msg("Failed to start metrics server")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Info().Msg("Shutting down application")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := metricsServer.Stop(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to stop metrics server")
	}

	os.Exit(0)
}
