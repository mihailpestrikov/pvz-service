package main

import (
	"avito-backend-trainee-assignment-spring-2025/internal/api/handlers"
	"avito-backend-trainee-assignment-spring-2025/internal/repository/postgres"
	"avito-backend-trainee-assignment-spring-2025/internal/services"
	"avito-backend-trainee-assignment-spring-2025/pkg/config"
	"avito-backend-trainee-assignment-spring-2025/pkg/logger"
	"avito-backend-trainee-assignment-spring-2025/pkg/metrics"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
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

	userRepo := postgres.NewUserRepository(db)
	pvzRepo := postgres.NewPVZRepository(db)
	receptionRepo := postgres.NewReceptionRepository(db)
	productRepo := postgres.NewProductRepository(db)
	txManager := postgres.NewTxManager(db)

	userService := services.NewUserService(userRepo, cfg.JWT, txManager)
	pvzService := services.NewPVZService(pvzRepo, txManager)
	receptionService := services.NewReceptionService(receptionRepo, pvzRepo, txManager)
	productService := services.NewProductService(productRepo, receptionRepo, txManager)

	handler := handlers.NewHandler(
		userService,
		pvzService,
		receptionService,
		productService,
		cfg,
	)

	router := handler.InitRoutes()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	metricsServer := metrics.NewServer(cfg.Prometheus.Port)
	go func() {
		if err := metricsServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error().Err(err).Msg("Failed to start metrics server")
		}
	}()

	go func() {
		log.Info().Str("address", server.Addr).Msg("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Failed to start HTTP server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	log.Info().Msg("Shutting down application")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to stop HTTP server")
	}
	log.Info().Msg("HTTP server stopped")

	if err := metricsServer.Stop(ctx); err != nil {
		log.Error().Err(err).Msg("Failed to stop metrics server")
	}
	log.Info().Msg("Metrics server stopped")

	log.Info().Msg("Application shutdown complete")
}
