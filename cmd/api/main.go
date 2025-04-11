package main

import (
	"avito-backend-trainee-assignment-spring-2025/pkg/config"
	"avito-backend-trainee-assignment-spring-2025/pkg/logger"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
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

	os.Exit(0)
}
