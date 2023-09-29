package cmd

import (
	"context"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/nightscout"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newDeleteCommand(ctx context.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:           "delete",
		PreRun:        preRun(),
		PostRun:       postRun(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		newDeleteTreatment(ctx),
	)

	return cmd
}

func newDeleteTreatment(ctx context.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:           "treatment ID [...]",
		Short:         "delete a treatment",
		Hidden:        true,
		PreRun:        preRun(),
		PostRun:       postRun(),
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := settings.LoadConfig(); err != nil {
				return errors.Wrap(err, "cant load config")
			}

			if err := validateConfig(); err != nil {
				return err
			}

			jwtToken, err := nightscout.NewJWTToken(settings.Nightscout().URL, settings.Nightscout().APIToken)
			if err != nil {
				return err
			}

			ns, err := nightscout.NewWithJWTToken(settings.Nightscout().URL, jwtToken)
			if err != nil {
				return err
			}

			for i := 0; i < len(args); i++ {
				err = ns.Treatments().Delete(ctx, args[i])
				if err != nil {
					return err
				}

				log.Info().Msg("OK")

			}

			return nil
		},
	}

	return cmd
}
