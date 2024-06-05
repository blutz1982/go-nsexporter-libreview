package cmd

import (
	"context"
	"os"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/nightscout"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/printer"
	"github.com/spf13/cobra"
)

func newListCommand(ctx context.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:           "list",
		Hidden:        true,
		PreRun:        preRun(),
		PostRun:       postRun(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		newListTreatment(ctx),
		newListDeviceStatus(ctx),
		newListGlucose(ctx),
	)

	return cmd
}

func newListDeviceStatus(ctx context.Context) *cobra.Command {

	var (
		kind string
	)

	cmd := &cobra.Command{
		Use:           "devices",
		PreRun:        preRun(),
		PostRun:       postRun(),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			dateFrom, dateTo, err := settings.DateRange()
			if err != nil {
				return err
			}

			ns, err := getNightscoutClient(ctx)
			if err != nil {
				return err
			}

			ds, err := ns.DeviceStatus().List(ctx, nightscout.ListOptions{
				DateFrom: dateFrom,
				DateTo:   dateTo,
				Count:    settings.NightscoutMaxEnties(),
				Kind:     kind,
			})
			if err != nil {
				return err
			}

			return printer.NewPrinter(settings.OutFormat(), os.Stdout).Print(ds)
		},
	}
	fs := cmd.Flags()
	settings.AddListFlags(fs)
	fs.StringVar(&kind, "device-type", "", "device type (e.g. BRIDGE or PHONE)")

	return cmd

}

func newListTreatment(ctx context.Context) *cobra.Command {

	var (
		kind string
	)

	cmd := &cobra.Command{
		Use:           "treatments",
		PreRun:        preRun(),
		PostRun:       postRun(),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			dateFrom, dateTo, err := settings.DateRange()
			if err != nil {
				return err
			}

			ns, err := getNightscoutClient(ctx)
			if err != nil {
				return err
			}

			t, err := ns.Treatments().List(ctx, nightscout.ListOptions{
				Kind:     kind,
				DateFrom: dateFrom,
				DateTo:   dateTo,
				Count:    settings.NightscoutMaxEnties(),
			})
			if err != nil {
				return err
			}

			return printer.NewPrinter(settings.OutFormat(), os.Stdout).Print(t)
		},
	}
	fs := cmd.Flags()
	settings.AddListFlags(fs)
	fs.StringVar(&kind, "kind", "", "kind of treatments")

	return cmd
}

func newListGlucose(ctx context.Context) *cobra.Command {

	var (
		kind string
	)

	cmd := &cobra.Command{
		Use:           "glucose",
		PreRun:        preRun(),
		PostRun:       postRun(),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			dateFrom, dateTo, err := settings.DateRange()
			if err != nil {
				return err
			}

			ns, err := getNightscoutClient(ctx)
			if err != nil {
				return err
			}

			t, err := ns.Glucose().List(ctx, nightscout.ListOptions{
				Kind:     kind,
				DateFrom: dateFrom,
				DateTo:   dateTo,
				Count:    settings.NightscoutMaxEnties(),
			})
			if err != nil {
				return err
			}

			return printer.NewPrinter(settings.OutFormat(), os.Stdout).Print(t)
		},
	}
	fs := cmd.Flags()
	settings.AddListFlags(fs)
	fs.StringVar(&kind, "kind", nightscout.Sgv, "type of entries")

	return cmd
}
