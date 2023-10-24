package cmd

import (
	"context"
	"os"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/nightscout"
	"github.com/blutz1982/go-nsexporter-libreview/pkg/nsgraph"
	"github.com/spf13/cobra"
)

const (
	defaultTargetLow  = 3.9
	defaultTargetHigh = 12.6
)

func newGraphommand(ctx context.Context) *cobra.Command {

	var (
		filePath string
	)

	cmd := &cobra.Command{
		Use:           "graph",
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

			p, err := ns.Profiles().Get(ctx)
			if err != nil {
				return err
			}

			store := p.Store[p.DefaultProfile]

			targetLow := defaultTargetLow
			targetHigh := defaultTargetHigh

			if len(store.TargetHigh) > 0 {
				targetHigh = store.TargetHigh[0].Value
			}

			if len(store.TargetLow) > 0 {
				targetLow = store.TargetLow[0].Value
			}

			entries, err := ns.Glucose().List(ctx, nightscout.ListOptions{
				Kind:     nightscout.Sgv,
				DateFrom: dateFrom,
				DateTo:   dateTo,
				Count:    settings.NightscoutMaxEnties(),
			})
			if err != nil {
				return err
			}

			f, err := os.Create(filePath)
			if err != nil {
				return err
			}
			defer f.Close()

			return nsgraph.DrawChart(entries, f, targetLow, targetHigh)

		},
	}

	fs := cmd.Flags()
	settings.AddListFlags(fs)
	fs.StringVar(&filePath, "filename", "svg.png", "path to file")

	return cmd
}
