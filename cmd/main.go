package main

import (
	"context"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "2006/02/01 15:04:05"})
	log.Info().Msg("Starting app")

	ctx := context.Background()

	config, err := pgxpool.ParseConfig(os.Getenv("PG_HOMEBUDGET_DB")) // DatabaseURL
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to parse conn string (%s): %w", os.Getenv("PG_HOMEBUDGET_DB"), err)
	}

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		log.Fatal().Err(err).Msgf("unable to connect to database: %w", err)
	}
	defer pool.Close()

	app := &app{
		Repository: NewRepository(pool),
	}

	app.Serve(ctx)
}
