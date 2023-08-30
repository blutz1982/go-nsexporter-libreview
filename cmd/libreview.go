package cmd

import (
	"context"
	"time"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/libreview"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/nightscout"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/transform"
	"github.com/spf13/cobra"
)

func newLibreCommand(ctx context.Context) *cobra.Command {

	var (
		fromDate    string
		toDate      string
		minInterval int
	)

	cmd := &cobra.Command{
		Use:           "libreview",
		Short:         "export data to libreview",
		PreRun:        preRun(),
		PostRun:       postRun(),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			ns, err := nightscout.NewWithConfig(settings.Nightscout())
			if err != nil {
				return err
			}

			lvges, err := getLibreviewGlucoseEntriesFromNightscout(ns, fromDate, toDate, minInterval)
			if err != nil {
				return err
			}

			lv := libreview.NewWithConfig(settings.Libreview())

			if err := lv.Auth(); err != nil {
				return err
			}

			return lv.ImportMeasurements(
				libreview.WithGlucoseEntries(lvges),
			)

		},
	}

	fs := cmd.Flags()
	fs.StringVar(&fromDate, "date-from", "", "start of sampling period")
	fs.StringVar(&toDate, "date-to", "", "end of sampling period")
	fs.IntVar(&minInterval, "min-interval", 12, "filter: min sample interval (minutes)")
	return cmd
}

func getLibreviewGlucoseEntriesFromNightscout(ns nightscout.Client, fromDate, toDate string, minInterval int) (libreview.GlucoseEntries, error) {

	var entries libreview.GlucoseEntries

	if len(fromDate) == 0 {
		fromDate = time.Now().Format(nightscout.TimestampLayout)
	}

	if len(toDate) == 0 {
		toDate = time.Now().AddDate(0, 0, 1).Format(nightscout.TimestampLayout)
	}

	nsGlucoseEntries, err := ns.GetGlucoseEntries(fromDate, toDate, nightscout.MaxEnties)
	if err != nil {
		return entries, err
	}

	nsGlucoseEntries = nsGlucoseEntries.Downsample(minInterval)

	nsGlucoseEntries.Visit(func(e *nightscout.GlucoseEntry, err error) error {
		entries = append(entries, transform.NSToLibreGlucoseEntry(e))
		return nil
	})

	return entries, nil

}
