package cmd

import (
	"context"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/env"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var settings = env.Default

type cobraRunFunc func(cmd *cobra.Command, args []string)

func preRun() cobraRunFunc {
	return func(cmd *cobra.Command, args []string) {

		logLevel := zerolog.InfoLevel
		if settings.Debug {
			logLevel = zerolog.DebugLevel
		}
		zerolog.SetGlobalLevel(logLevel)

		debug("app started. Version %s", settings.Version())
	}
}

func postRun() cobraRunFunc {
	return func(cmd *cobra.Command, args []string) {
		debug("app done")
	}
}

func fatal(err error) {
	log.Fatal().Err(err).Send()
}

func debug(format string, v ...interface{}) {
	log.Debug().Msgf(format, v...)
}

func info(format string, v ...interface{}) {
	log.Info().Msgf(format, v...)
}

func NewRootCmd(ctx context.Context, args []string) *cobra.Command {

	cmd := &cobra.Command{
		Use:          "nsexport",
		Short:        "nightscout exporter",
		Version:      settings.Version(),
		PreRun:       preRun(),
		PostRun:      postRun(),
		SilenceUsage: true,
	}
	fs := cmd.PersistentFlags()
	settings.AddCommonFlags(fs)

	cobra.OnInitialize(func() {
		if err := settings.LoadConfig(); err != nil {
			log.Fatal().Err(err).Msg("cant load config")
		}
	})

	cmd.AddCommand(
		newLibreCommand(ctx),
	)

	return cmd

}
