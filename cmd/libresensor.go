package cmd

import (
	"context"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/libreview"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newLibreNewSensor(ctx context.Context) *cobra.Command {

	var (
		setDevice bool
	)

	cmd := &cobra.Command{
		Use:           "libre-new-sensor [SERIAL]",
		Hidden:        true,
		PreRun:        preRun(),
		PostRun:       postRun(),
		Args:          cobra.MinimumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			serial := args[0]

			if err := settings.LoadConfig(); err != nil {
				return errors.Wrap(err, "cant load config")
			}

			lv, err := libreview.NewWithConfig(settings.Libreview())
			if err != nil {
				return err
			}

			if err := lv.Auth(setDevice); err != nil {
				return err
			}

			return lv.NewSensor(serial)
		},
	}

	fs := cmd.Flags()

	settings.AddListFlags(fs)

	fs.BoolVar(&setDevice, "set-device", true, "Set this app as main user device. Necessary if the main device was set by another application (e.g. Librelink)")

	return cmd
}
