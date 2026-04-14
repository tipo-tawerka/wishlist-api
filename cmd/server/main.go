package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/tipo-tawerka/wishlist-api/config"
	"github.com/tipo-tawerka/wishlist-api/internal/app"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.NewConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	if err := app.Run(cfg, &logger); err != nil {
		logger.Fatal().Err(err).Msg("app stopped with error")
	}
}
