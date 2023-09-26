package nightscout

import (
	"context"
	"fmt"
	"strconv"
	"time"
)

type TreatmentsGetter interface {
	Treatments(kind string) TreatmentInterface
}

type TreatmentInterface interface {
	Get(ctx context.Context, opts GetOptions) (*Treatments, error)
}

type treatments struct {
	client RestInterface
	kind   string
}

func (t treatments) Get(ctx context.Context, opts GetOptions) (result *Treatments, err error) {
	result = &Treatments{}
	r := t.client.Get().
		Name("treatments").
		Param("find[created_at][$gte]", opts.DateFrom.UTC().Format(time.RFC3339)).
		Param("find[created_at][$lte]", opts.DateTo.UTC().Format(time.RFC3339)).
		Param(fmt.Sprintf("find[%s][$gt]", t.kind), "0").
		Param("count", strconv.Itoa(opts.Count))

	if len(opts.URLToken) > 0 {
		r = r.Param("token", opts.URLToken)
	}

	err = r.Do(ctx).Into(result)

	return

}

// newTreatments returns a Treatments
func newTreatments(c Client, kind string) *treatments {
	return &treatments{
		client: c.RESTClient(),
		kind:   kind,
	}
}

type Treatment struct {
	ID        string    `json:"_id"`
	CreatedAt time.Time `json:"created_at"`
	Insulin   float64   `json:"insulin"`
	Carbs     float64   `json:"carbs"`
}

type Treatments []*Treatment

type TreatmentsVisitorFunc func(*Treatment, error) error

func (es Treatments) Visit(fn TreatmentsVisitorFunc) error {
	var err error
	for _, entry := range es {
		if err = fn(entry, err); err != nil {
			return err
		}
	}
	return nil
}

func (t Treatments) Len() int {
	return len(t)
}

func (r *Treatments) Append(e *Treatment) {
	*r = append(*r, e)
}

type TreatmentFilterFunc func(*Treatment) bool

func TreatmentOnlyAfter(date time.Time) TreatmentFilterFunc {
	return func(e *Treatment) bool {
		return e.CreatedAt.After(date)
	}
}

func (es *Treatments) Filter(fn TreatmentFilterFunc) (result *Treatments) {
	result = &Treatments{}
	es.Visit(func(e *Treatment, _ error) error {
		if fn(e) {
			result.Append(e)
		}
		return nil
	})
	return result
}
