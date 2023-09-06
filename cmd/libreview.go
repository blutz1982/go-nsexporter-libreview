package cmd

import (
	"context"
	"math/rand"
	"time"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/libreview"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/nightscout"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/transform"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newLibreCommand(ctx context.Context) *cobra.Command {

	const frequencyDeflectionPercent int = 30

	var (
		fromDate         string
		toDate           string
		minInterval      int
		dryRun           bool
		avgScanFrequency int
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

			nsGlucoseEntries, err := getNSGlucoseEntries(ns, fromDate, toDate)
			if err != nil {
				return err
			}

			nsGlucoseEntries = nsGlucoseEntries.Downsample(nightscout.DownsampleMinutes(minInterval))

			log.Info().
				Int("count", nsGlucoseEntries.Len()).
				Msg("Get glucose entries from Nightscout")

			var libreScheduledGlucoseEntries libreview.ScheduledGlucoseEntries
			nsGlucoseEntries.Visit(func(e *nightscout.GlucoseEntry, err error) error {
				libreScheduledGlucoseEntries.Append(transform.NSToLibreScheduledGlucoseEntry(e))
				log.Debug().
					Time("ts", e.DateString.Local()).
					Float64("svg", e.Sgv.Float64()).
					Str("direction", e.Direction).
					Msg("entry")
				return nil
			})

			min, max := getRangeSpread(avgScanFrequency, frequencyDeflectionPercent)

			nsGlucoseEntriesUnscheduled := nsGlucoseEntries.Downsample(func() float64 {
				interval := float64(rand.Intn(max-min) + min)
				return interval

			})

			var libreUnscheduledGlucoseEntries libreview.UnscheduledContinuousGlucoseEntries
			nsGlucoseEntriesUnscheduled.Visit(func(e *nightscout.GlucoseEntry, err error) error {
				// fmt.Println("date, direction ", e.DateString, e.Direction)
				libreUnscheduledGlucoseEntries.Append(transform.NSToLibreUnscheduledGlucoseEntry(e))
				return nil

			})

			for _, e := range libreUnscheduledGlucoseEntries {
				log.Debug().
					Time("ts", e.Timestamp).
					Float64("svg", e.ValueInMgPerDl).
					Str("direction", e.ExtendedProperties.TrendArrow).
					Msg("Unscheduled Glucose entry")
			}

			lv := libreview.NewWithConfig(settings.Libreview())

			if err := lv.Auth(); err != nil {
				return err
			}

			if dryRun {
				log.Info().Msg("dry run mode. Nothing to post")
			}

			if err := lv.ImportMeasurements(
				dryRun,
				libreview.WithScheduledGlucoseEntries(libreScheduledGlucoseEntries),
				libreview.WithUnscheduledGlucoseEntries(libreUnscheduledGlucoseEntries),
			); err != nil {
				return err
			}

			log.Info().Int("count", nsGlucoseEntries.Len()).Msg("Export success")

			return nil

		},
	}

	fs := cmd.Flags()
	fs.StringVar(&fromDate, "date-from", "", "start of sampling period")
	fs.StringVar(&toDate, "date-to", "", "end of sampling period")
	fs.IntVar(&minInterval, "min-interval", 12, "filter: min sample interval (minutes)")
	fs.IntVar(&avgScanFrequency, "scan-frequency", 90, "average scan frequency (minutes). e.g. scan internal min=avg-30%, max=avg+30%")
	fs.BoolVar(&dryRun, "dry-run", false, "dont post measurement to libreview")
	return cmd
}

func Percent(percent int, all int) float64 {
	return ((float64(all) * float64(percent)) / float64(100))
}

func getRangeSpread(avgVal, percentSpread int) (min, max int) {
	return int(float64(avgVal) - Percent(percentSpread, avgVal)),
		int(float64(avgVal) + Percent(percentSpread, avgVal))
}

func getNSGlucoseEntries(ns nightscout.Client, fromDateStr, toDateStr string) (nightscout.GlucoseEntries, error) {

	var (
		fromDate time.Time
		toDate   time.Time
		err      error
	)

	if len(fromDateStr) == 0 {
		fromDate = time.Now()
	} else {
		fromDate, err = time.Parse(nightscout.TimestampLayout, fromDateStr)
		if err != nil {
			return nil, err
		}
	}

	if len(toDateStr) == 0 {
		toDate = fromDate.AddDate(0, 0, 1)
	} else {
		toDate, err = time.Parse(nightscout.TimestampLayout, toDateStr)
		if err != nil {
			return nil, err
		}
	}

	return ns.GetGlucoseEntries(fromDate, toDate, nightscout.MaxEnties)

}
