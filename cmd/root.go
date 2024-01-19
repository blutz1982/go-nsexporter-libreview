package cmd

import (
	"context"
	"time"

	"github.com/blutz1982/go-nsexporter-libreview/internal/version"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/env"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/nightscout"
	"github.com/pkg/errors"
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

		info("app started. %s", version.FormatVersion())
	}
}

func postRun() cobraRunFunc {
	return func(cmd *cobra.Command, args []string) {
		info("app done")
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
		Use:          settings.AppName(),
		Short:        "nightscout exporter",
		Version:      version.FormatVersion(),
		PreRun:       preRun(),
		PostRun:      postRun(),
		SilenceUsage: true,
	}

	cmd.SetVersionTemplate(version.CmdVersionTemplate)

	fs := cmd.PersistentFlags()
	settings.AddCommonFlags(fs)

	cobra.OnInitialize(func() {

		if len(settings.Timezone) == 0 {
			return
		}

		loc, err := time.LoadLocation(settings.Timezone)
		if err != nil {
			log.Fatal().Err(err).Msg("cant load timezone")
		}

		time.Local = loc

	})

	cmd.AddCommand(
		newLibreCommand(ctx),
		newConfigCommand(ctx),
		newCreateCommand(ctx),
		newDeleteCommand(ctx),
		newListCommand(ctx),
		newGraphommand(ctx),
		newLibreAuth(ctx),
		newLibreNewSensor(ctx),
	)

	return cmd

}

func getNightscoutClient(ctx context.Context) (nightscout.Client, error) {
	if err := settings.LoadConfig(); err != nil {
		return nil, errors.Wrap(err, "cant load config")
	}

	jwtToken, err := nightscout.NewJWTToken(ctx, settings.Nightscout().URL, settings.Nightscout().APIToken)
	if err != nil {
		return nil, err
	}

	return nightscout.NewWithJWTToken(settings.Nightscout().URL, jwtToken)
}
