package cmd

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/libreview"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/nightscout"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/transform"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

func newLibreCommand(ctx context.Context) *cobra.Command {

	const (
		frequencyDeflectionPercent int = 30
	)

	var (
		minInterval       string
		dryRun            bool
		avgScanFrequency  int
		setDevice         bool
		lastTimestampFile string
		measurements      []string
		token             string
		newSensorSerial   string
	)

	cmd := &cobra.Command{
		Use:           "libreview",
		Short:         "export data to libreview",
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

			nsInsulinEntries, err := ns.Treatments().List(ctx, nightscout.ListOptions{
				Kind:     nightscout.Insulin,
				DateFrom: dateFrom,
				DateTo:   dateTo,
				Count:    settings.NightscoutMaxEnties(),
			})
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
					Float64("insulin", t.Insulin).
					Str("type", transform.LongActingInsulinMap[t.InsulinInjections.IsLongActing()]).
					Msg("Insulin entry")
				return nil
			})

			nsCarbsEntries, err := ns.Treatments().List(ctx, nightscout.ListOptions{
				Kind:     nightscout.Carbs,
				DateFrom: dateFrom,
				DateTo:   dateTo,
				Count:    settings.NightscoutMaxEnties(),
			})
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
					Float64("carbs", t.Carbs).
					Msg("Food entry")
				return nil
			})

			nsGlucoseEntries, err := ns.Glucose().List(ctx, nightscout.ListOptions{
				DateFrom: dateFrom,
				DateTo:   dateTo,
				Count:    settings.NightscoutMaxEnties(),
				Kind:     nightscout.Sgv,
			})
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
					Time("ts", e.Date.Time().Local()).
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

			var libreGenericEntries libreview.GenericEntries
			lastScan, ok := libreUnscheduledGlucoseEntries.Last()
			if ok {
				libreGenericEntries.Append(transform.LibreUnscheduledContinuousGlucoseEntryToSensorStart(lastScan))
			}

			measurementMap := map[string]libreview.MeasuremenModificator{
				"scheduledContinuousGlucose":   libreview.WithScheduledGlucoseEntries(libreScheduledGlucoseEntries),
				"unscheduledContinuousGlucose": libreview.WithUnscheduledGlucoseEntries(libreUnscheduledGlucoseEntries),
				"insulin":                      libreview.WithInsulinEntries(libreInsulinEntries),
				"food":                         libreview.WithFoodEntries(libreFoodEntries),
				"generic":                      libreview.WithGenericEntries(libreGenericEntries),
			}

			var modificators []libreview.MeasuremenModificator

			importGeneric := false
			for _, m := range measurements {
				modificator, ok := measurementMap[m]
				if ok {
					modificators = append(modificators, modificator)
				}
				if m == "generic" {
					importGeneric = len(newSensorSerial) > 0
				}
			}

			if importGeneric {
				log.Info().
					Str("serial", newSensorSerial).
					Time("install time", lastScan.Timestamp).
					Msg("Prepare sensor start generic entry")
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
				Msg("use token for libreview")

			resp, err := lv.ImportMeasurements(modificators...)
			if err != nil {
				return err
			}

			if len(libreGenericEntries) > 0 && importGeneric {
				err := lv.NewSensor(newSensorSerial)
				if err != nil {
					log.Error().
						Err(err).
						Msg("Posible new sensor install failed")
				}
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

	settings.AddListFlags(fs)

	fs.StringVar(&minInterval, "min-interval", "10m10s", "Filter: minimum sample interval (duration)")
	fs.IntVar(&avgScanFrequency, "scan-frequency", 90, "Average scan frequency (minutes). e.g. scan internal min=avg-30%, max=avg+30%")
	fs.BoolVar(&dryRun, "dry-run", false, "Do not post measurement to LibreView")
	fs.BoolVar(&setDevice, "set-device", true, "Set this app as main user device. Necessary if the main device was set by another application (e.g. Librelink)")
	fs.StringVar(&lastTimestampFile, "last-ts-file", "", "Path to last timestamp file (for example ./last.ts )")
	fs.StringSliceVar(&measurements, "measurements", libreview.AllMeasurements, "measurements to upload")
	fs.StringVar(&token, "token", "", "use existing libreview token (beta)")
	fs.StringVar(&newSensorSerial, "install-new-sensor-sn", "", "new sensor serial number")

	err := fs.MarkHidden("token")
	if err != nil {
		panic(err)
	}

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
