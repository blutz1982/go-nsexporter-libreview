package cmd

import (
	"context"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/libreview"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newLibreAuth(ctx context.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:           "libreauth",
		Hidden:        true,
		PreRun:        preRun(),
		PostRun:       postRun(),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := settings.LoadConfig(); err != nil {
				return errors.Wrap(err, "cant load config")
			}

			lv, err := libreview.NewWithConfig(settings.Libreview())
			if err != nil {
				return err
			}

			if err := lv.Auth(false); err != nil {
				return err
			}

			return nil
		},
	}

	return cmd
}
