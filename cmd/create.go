package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/nightscout"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

const formatCreate = `Created.

ID: %s
CreatedAt: %s
EventType: %s
EnteredBy: %s
Insulin: %.1f
InsulinInjections: %s
Carbs: %.1f
`

func newCreateCommand(ctx context.Context) *cobra.Command {

	cmd := &cobra.Command{
		Use:           "create",
		Hidden:        true,
		PreRun:        preRun(),
		PostRun:       postRun(),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		newCreateTreatment(ctx),
	)

	return cmd
}

func newCreateTreatment(ctx context.Context) *cobra.Command {

	var (
		insulin       float64
		insulinType   string
		carbs         float64
		createTime    string
		treatmentType string
		enteredBy     string
	)

	cmd := &cobra.Command{
		Use:           "treatment",
		PreRun:        preRun(),
		PostRun:       postRun(),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {

			if insulin == 0 && carbs == 0 {
				return errors.New("nothing to create")
			}

			t := nightscout.NewTreatment()

			t.Carbs = carbs
			t.Insulin = insulin

			insType, err := nightscout.ParseInsulinType(insulinType)
			if err != nil {
				return err
			}

			if insType.IsLongActing() {
				t.InsulinInjections = nightscout.NewInsulinInjections(insulin, insType)
			}

			t.EnteredBy = enteredBy
			t.EventType = treatmentType

			if len(createTime) > 0 {
				ts, err := time.Parse(time.RFC3339, createTime)
				if err != nil {
					return err
				}
				t.CreatedAt = ts.UTC()

			} else {
				t.CreatedAt = time.Now().UTC()
			}

			ns, err := getNightscoutClient()
			if err != nil {
				return err
			}

			created, err := ns.Treatments().Create(ctx, t)

			created.Visit(func(t *nightscout.Treatment, err error) error {

				fmt.Printf(formatCreate,
					t.ID,
					t.CreatedAt.Local().String(),
					t.EventType,
					t.EnteredBy,
					t.Insulin,
					t.InsulinInjections.String(),
					t.Carbs,
				)

				return nil
			})

			return nil
		},
	}
	fs := cmd.Flags()
	fs.StringVar(&insulinType, "insulin-type", nightscout.Fiasp.String(), "insulin type")
	fs.StringVar(&enteredBy, "entered-by", "nsexport", "entered by")
	fs.StringVar(&treatmentType, "treatment-type", "", "treatment type")
	fs.Float64Var(&insulin, "insulin", 0, "insulin units")
	fs.Float64Var(&carbs, "carbs", 0, "carbs units")
	fs.StringVar(&createTime, "ts", "", "entry create timestamp (default - current time)")

	return cmd
}
