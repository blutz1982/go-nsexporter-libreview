package nightscout

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/rest"
)

type TreatmentsGetter interface {
	Treatments() TreatmentInterface
}

type TreatmentInterface interface {
	Get(ctx context.Context, opts GetOptions) (*Treatments, error)
	Create(ctx context.Context, treatment *Treatment) (result Treatments, err error)
	Delete(ctx context.Context, id string) error
}

type treatments struct {
	client rest.Interface
}

func (t *treatments) Get(ctx context.Context, opts GetOptions) (result *Treatments, err error) {
	result = &Treatments{}
	r := t.client.Get().
		Name("treatments").
		Param("find[created_at][$gte]", opts.DateFrom.UTC().Format(time.RFC3339)).
		Param("find[created_at][$lte]", opts.DateTo.UTC().Format(time.RFC3339)).
		Param(fmt.Sprintf("find[%s][$gt]", opts.Kind), "0").
		Param("count", strconv.Itoa(opts.Count))

	err = r.Do(ctx).Into(result)

	return

}

func (t *treatments) Create(ctx context.Context, treatment *Treatment) (result Treatments, err error) {

	err = t.client.Post().
		Name("treatments").
		Body(treatment).
		Do(ctx).
		Into(&result)

	return

}

func (t *treatments) Delete(ctx context.Context, id string) error {
	return t.client.Delete().
		Resource("treatments").
		Name(id).
		Do(ctx).
		Error()
}

// newTreatments returns a Treatments
func newTreatments(c Client) *treatments {
	return &treatments{
		client: c.RESTClient(),
	}
}

type InsulinInjections string

var reLongActingInsulin = regexp.MustCompile(`^.*(Lantus|Toujeo|Tresiba).*$`)

func (ii InsulinInjections) IsLongActing() bool {
	return reLongActingInsulin.MatchString(ii.String())
}

func (ii InsulinInjections) String() string {
	return string(ii)
}

func NewInsulinInjections(units float64, insType InsulinType) InsulinInjections {
	return InsulinInjections(fmt.Sprintf("[{\"insulin\":\"%s\",\"units\":%.1f}]", insType.String(), units))
}

type Treatment struct {
	ID                string            `json:"_id"`
	EventType         string            `json:"eventType"`
	EnteredBy         string            `json:"enteredBy"`
	CreatedAt         time.Time         `json:"created_at"`
	Insulin           float64           `json:"insulin"`
	Carbs             float64           `json:"carbs"`
	InsulinInjections InsulinInjections `json:"insulinInjections"`
}

func (t *Treatment) MarshalJSON() ([]byte, error) {
	// exclude ID
	return json.Marshal(&struct {
		EventType         string            `json:"eventType"`
		EnteredBy         string            `json:"enteredBy"`
		CreatedAt         time.Time         `json:"created_at"`
		Insulin           float64           `json:"insulin"`
		Carbs             float64           `json:"carbs"`
		InsulinInjections InsulinInjections `json:"insulinInjections"`
	}{
		EventType:         t.EventType,
		EnteredBy:         t.EnteredBy,
		CreatedAt:         t.CreatedAt,
		Insulin:           t.Insulin,
		Carbs:             t.Carbs,
		InsulinInjections: t.InsulinInjections,
	})
}

func (t *Treatment) Kind() string {
	return "Treatment"
}

type InsulinType int8

const (
	Fiasp InsulinType = iota
	Novorapid
	Humalog
	Lispro
	Actapid
	Lantus
	Toujeo
	InsulinTypeUnknown
)

var insulinTypeLoCaseMap = map[string]InsulinType{
	"fiasp":     Fiasp,
	"novorapid": Novorapid,
	"humalog":   Humalog,
	"lispro":    Lispro,
	"actapid":   Actapid,
	"lantus":    Lantus,
	"toujeo":    Toujeo,
}

func ParseInsulinType(value string) (InsulinType, error) {
	t, ok := insulinTypeLoCaseMap[strings.ToLower(value)]
	if !ok {
		return InsulinTypeUnknown, fmt.Errorf("parse error: unknown insulin type %s", value)
	}
	return t, nil
}

func (it InsulinType) String() string {
	switch it {
	case Lantus:
		return "Lantus"
	case Toujeo:
		return "Toujeo"
	case Fiasp:
		return "Fiasp"
	case Novorapid:
		return "Novorapid"
	case Humalog:
		return "Humalog"
	case Lispro:
		return "Lispro"
	case Actapid:
		return "Actapid"
	default:
		return "unknown"
	}
}

func (it InsulinType) IsLongActing() bool {
	switch it {
	case Lantus, Toujeo:
		return true
	default:
		return false
	}
}

func NewTreatment() *Treatment {
	return new(Treatment)
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
