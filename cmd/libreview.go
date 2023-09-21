package cmd

import (
	"context"
	"errors"
	"math/rand"
	"os"
	"time"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/libreview"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/nightscout"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/transform"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func validateConfig() error {
	if settings.Nightscout() == nil {
		return errors.New("bad config. nightscout section not found")
	}

	if settings.Libreview() == nil {
		return errors.New("bad config. libreview section not found")
	}

	return nil

}

func newLibreCommand(ctx context.Context) *cobra.Command {

	const (
		frequencyDeflectionPercent int = 30
		defaultTSLayout                = "2006-01-02"
	)

	var (
		fromDate          string
		toDate            string
		dateOffset        string
		minInterval       string
		dryRun            bool
		avgScanFrequency  int
		setDevice         bool
		tsLayout          string
		lastTimestampFile string
		measurements      []string
		token             string
	)

	cmd := &cobra.Command{
		Use:           "libreview",
		Short:         "export data to libreview",
		PreRun:        preRun(),
		PostRun:       postRun(),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			if err := validateConfig(); err != nil {
				return err
			}

			dateFrom, dateTo, err := getDateRange(fromDate, toDate, dateOffset, tsLayout)
			if err != nil {
				return err
			}

			var lastTS *time.Time

			if len(lastTimestampFile) > 0 {
				lastTS, err = getLastTS(lastTimestampFile)
				if err != nil {
					return err
				}
			}

			opts := nightscout.GetOptions{
				DateFrom: dateFrom,
				DateTo:   dateTo,
				Count:    nightscout.MaxEnties,
				APIToken: settings.Nightscout().APIToken,
			}

			ns, err := nightscout.New(settings.Nightscout().URL)
			if err != nil {
				return err
			}

			nsInsulinEntries, err := ns.Treatments(nightscout.Insulin).Get(ctx, opts)
			if err != nil {
				return err
			}

			if lastTS != nil {
				nsInsulinEntries = nsInsulinEntries.Filter(nightscout.TreatmentOnlyAfter(lastTS.UTC().Add(time.Minute)))
			}

			log.Info().
				Int("count", nsInsulinEntries.Len()).
				Time("fromDate", dateFrom).
				Time("toDate", dateTo).
				Msg("Get insulin entries from Nightscout")

			var libreInsulinEntries libreview.InsulinEntries

			nsInsulinEntries.Visit(func(t *nightscout.Treatment, _ error) error {
				libreInsulinEntries.Append(transform.NSToLibreInsulinEntry(t))

				log.Debug().
					Time("ts", t.CreatedAt.Local()).
					Int("insulin", t.Insulin).
					Msg("Insulin entry")
				return nil
			})

			nsCarbsEntries, err := ns.Treatments(nightscout.Carbs).Get(ctx, opts)
			if err != nil {
				return err
			}

			if lastTS != nil {
				nsCarbsEntries = nsCarbsEntries.Filter(nightscout.TreatmentOnlyAfter(lastTS.UTC().Add(time.Minute)))
			}

			log.Info().
				Int("count", nsCarbsEntries.Len()).
				Time("fromDate", dateFrom).
				Time("toDate", dateTo).
				Msg("Get food entries from Nightscout")

			var libreFoodEntries libreview.FoodEntries

			nsCarbsEntries.Visit(func(t *nightscout.Treatment, err error) error {
				libreFoodEntries.Append(transform.NSToLibreFoodEntry(t))
				log.Debug().
					Time("ts", t.CreatedAt.Local()).
					Int("carbs", t.Carbs).
					Msg("Food entry")
				return nil
			})

			nsGlucoseEntries, err := ns.Glucose().Get(ctx, opts)
			if err != nil {
				return err
			}

			if lastTS != nil {
				nsGlucoseEntries = nsGlucoseEntries.Filter(nightscout.OnlyAfter(lastTS.UTC().Add(time.Minute)))
			}

			d, err := time.ParseDuration(minInterval)
			if err != nil {
				return err
			}

			nsGlucoseEntries = nsGlucoseEntries.Downsample(nightscout.DownsampleDuration(d))

			log.Info().
				Int("count", nsGlucoseEntries.Len()).
				Time("fromDate", dateFrom).
				Time("toDate", dateTo).
				Msg("Get scheduled glucose entries from Nightscout")

			var libreScheduledGlucoseEntries libreview.ScheduledContinuousGlucoseEntries
			nsGlucoseEntries.Visit(func(e *nightscout.GlucoseEntry, err error) error {
				libreScheduledGlucoseEntries.Append(transform.NSToLibreScheduledGlucoseEntry(e))
				log.Debug().
					Time("ts", e.DateString.Local()).
					Float64("svg", e.Sgv.Float64()).
					Str("direction", e.Direction).
					Msg("Scheduled Glucose entry")
				return nil
			})

			log.Info().
				Int("count", nsGlucoseEntries.Len()).
				Time("fromDate", dateFrom).
				Time("toDate", dateTo).
				Msg("Prepare unscheduled glucose entries")

			min, max := getRangeSpread(avgScanFrequency, frequencyDeflectionPercent)

			var libreUnscheduledGlucoseEntries libreview.UnscheduledContinuousGlucoseEntries

			nsGlucoseEntries.Downsample(func() time.Duration {
				return (time.Minute * time.Duration(rand.Intn(max-min)+min))

			}).Visit(func(e *nightscout.GlucoseEntry, _ error) error {
				libreUnscheduledGlucoseEntries.Append(transform.NSToLibreUnscheduledGlucoseEntry(e))
				return nil
			})

			libreUnscheduledGlucoseEntries.Visit(func(e *libreview.UnscheduledContinuousGlucoseEntry, _ error) error {
				log.Debug().
					Time("ts", e.Timestamp).
					Float64("svg", e.ValueInMgPerDl).
					Str("direction", e.ExtendedProperties.TrendArrow).
					Msg("Unscheduled Glucose entry")
				return nil
			})

			log.Info().
				Strs("measurements", measurements).
				Msg("Measurements to export")

			measurementMap := map[string]libreview.MeasuremenModificator{
				"scheduledContinuousGlucose":   libreview.WithScheduledGlucoseEntries(libreScheduledGlucoseEntries),
				"unscheduledContinuousGlucose": libreview.WithUnscheduledGlucoseEntries(libreUnscheduledGlucoseEntries),
				"insulin":                      libreview.WithInsulinEntries(libreInsulinEntries),
				"food":                         libreview.WithFoodEntries(libreFoodEntries),
			}

			var modificators []libreview.MeasuremenModificator

			for _, m := range measurements {
				modificator, ok := measurementMap[m]
				if ok {
					modificators = append(modificators, modificator)
				}
			}

			if dryRun || len(libreScheduledGlucoseEntries) == 0 || len(libreUnscheduledGlucoseEntries) == 0 || len(modificators) == 0 {
				log.Info().
					Bool("dry-run", dryRun).
					Msg("Nothing to post")
				return nil
			}

			lv, err := libreview.NewWithConfig(settings.Libreview())
			if err != nil {
				return err
			}

			if len(token) == 0 {
				if err := lv.Auth(setDevice); err != nil {
					return err
				}
			} else {
				lv.SetToken(token)
			}

			log.Debug().
				Str("token", lv.Token()).
				Msg("use token")

			resp, err := lv.ImportMeasurements(modificators...)
			if err != nil {
				return err
			}

			log.Info().
				Int("scheduledGlucoseEntries", resp.Result.MeasurementCounts.ScheduledGlucoseCount).
				Int("unscheduledGlucoseEntries", resp.Result.MeasurementCounts.UnScheduledGlucoseCount).
				Int("insulin", resp.Result.MeasurementCounts.InsulinCount).
				Int("food", resp.Result.MeasurementCounts.FoodCount).
				Msg("Export measurements success")

			lastTS = lv.LastImported()
			if lastTS != nil && len(lastTimestampFile) > 0 && !dryRun {
				if err := saveTS(lastTimestampFile, *lastTS); err != nil {
					return err
				}
				log.Info().
					Time("ts", *lastTS).
					Str("timestampFile", lastTimestampFile).
					Msg("Last scheduled glucose entry timestamp")
			}

			return nil

		},
	}

	fs := cmd.Flags()
	fs.StringVar(&tsLayout, "ts-layout", defaultTSLayout, "Timestamp layout for --date-from and --date-to flags. More https://go.dev/src/time/format.go")
	fs.StringVar(&fromDate, "date-from", "", "Start of sampling period")
	fs.StringVar(&dateOffset, "date-offset", "", "Start of sampling period with current time offset. Set in duration (e.g. 24h or 72h30m). Ignore --date-from and --date-to flags")
	fs.StringVar(&toDate, "date-to", "", "End of sampling period")
	fs.StringVar(&minInterval, "min-interval", "10m10s", "Filter: minimum sample interval (duration)")
	fs.IntVar(&avgScanFrequency, "scan-frequency", 90, "Average scan frequency (minutes). e.g. scan internal min=avg-30%, max=avg+30%")
	fs.BoolVar(&dryRun, "dry-run", false, "Do not post measurement to LibreView")
	fs.BoolVar(&setDevice, "set-device", true, "Set this app as main user device. Necessary if the main device was set by another application (e.g. Librelink)")
	fs.StringVar(&lastTimestampFile, "last-ts-file", "", "Path to last timestamp file (for example ./last.ts )")
	fs.StringSliceVar(&measurements, "measurements", libreview.AllMeasurements, "measurements to upload")
	fs.StringVar(&token, "token", "", "use existing token")

	return cmd
}

func saveTS(tsfile string, ts time.Time) error {
	return os.WriteFile(tsfile, []byte(ts.Format(time.RFC3339)), 0644)
}

func getLastTS(tsfile string) (*time.Time, error) {
	data, err := os.ReadFile(tsfile)
	if err != nil {
		return nil, nil
	}

	t, err := time.Parse(time.RFC3339, string(data))
	if err != nil {
		return nil, err
	}

	return &t, nil
}

func Percent(percent int, all int) float64 {
	return ((float64(all) * float64(percent)) / float64(100))
}

func getRangeSpread(avgVal, percentSpread int) (min, max int) {
	return int(float64(avgVal) - Percent(percentSpread, avgVal)),
		int(float64(avgVal) + Percent(percentSpread, avgVal))
}

func getDateRange(fromDateStr, toDateStr, dateOffset, tsLayout string) (fromDate, toDate time.Time, err error) {

	var duration time.Duration

	if len(dateOffset) > 0 {
		fromDateStr = ""
		toDateStr = ""
		duration, err = time.ParseDuration(dateOffset)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	if len(fromDateStr) == 0 {
		now := time.Now().Local()
		if len(dateOffset) > 0 {
			fromDate = now.Add(-duration)
		} else {
			fromDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
		}
	} else {
		fromDate, err = time.ParseInLocation(tsLayout, fromDateStr, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	if len(toDateStr) == 0 {
		toDate = time.Now().Local()
	} else {
		toDate, err = time.ParseInLocation(tsLayout, toDateStr, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, err
		}
	}

	return

}

// measurementMap := map[string]libreview.MeasuremenModificator{
// 	"scheduledContinuousGlucose":   libreview.WithScheduledGlucoseEntries(libreScheduledGlucoseEntries),
// 	"unscheduledContinuousGlucose": libreview.WithUnscheduledGlucoseEntries(libreUnscheduledGlucoseEntries),
// 	"insulin":                      libreview.WithInsulinEntries(libreInsulinEntries),
// }

// func getGlucoseEntries(ctx context.Context, ns nightscout.Client) libreview.MeasuremenModificator {

// }
