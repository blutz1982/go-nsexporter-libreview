package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/blutz1982/go-nsexporter-libreview/cmd"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/env"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	app string = "app"
)

func init() {

	env.Default.SetAppName(app)

	log.Logger = zerolog.New(zerolog.NewConsoleWriter(
		func(w *zerolog.ConsoleWriter) {
			w.Out = os.Stderr
			w.NoColor = false
			w.TimeFormat = time.RFC3339
		}),
	).With().Timestamp().Logger()

}

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	if err := cmd.NewRootCmd(ctx, os.Args[1:]).Execute(); err != nil {
		log.Fatal().Err(err).Msg("An error has accured")
	}
}
