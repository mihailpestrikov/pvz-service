package logger

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"avito-backend-trainee-assignment-spring-2025/pkg/config"
)

func Setup(cfg config.LoggerConfig) {
	level, err := zerolog.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		level = zerolog.DebugLevel
	}
	zerolog.SetGlobalLevel(level)

	var output io.Writer = os.Stdout

	if cfg.Output == "file" && cfg.FilePath != "" {
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err == nil {
			output = file
		}
	}

	var writer = output
	if cfg.Format == "pretty" {
		writer = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		}
	}

	logger := zerolog.New(writer).With().Timestamp()

	log.Logger = logger.Logger()

	log.Info().
		Str("level", level.String()).
		Str("format", cfg.Format).
		Str("output", cfg.Output).
		Msg("Logger initialized")
}
