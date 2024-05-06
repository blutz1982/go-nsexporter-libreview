package nightscout

import (
	"context"
	"strconv"
	"time"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/rest"
)

type GlucoseGetter interface {
	Glucose() GlucoseInterface
}

type GlucoseInterface interface {
	List(ctx context.Context, opts ListOptions) (*GlucoseEntries, error)
}

type glucose struct {
	client rest.Interface
}

// newGlucose returns a Glucos
func newGlucose(c Client) *glucose {
	return &glucose{
		client: c.RESTClient(),
	}
}

func (g glucose) List(ctx context.Context, opts ListOptions) (result *GlucoseEntries, err error) {
	result = &GlucoseEntries{}
	r := g.client.Get().
		Resource("entries").
		Name(opts.Kind).
		Param("find[date][$gte]", strconv.FormatInt(opts.DateFrom.UTC().UnixMilli(), 10)).
		Param("find[date][$lte]", strconv.FormatInt(opts.DateTo.UTC().UnixMilli(), 10)).
		Param("count", strconv.Itoa(opts.Count))

	err = r.Do(ctx).Into(result)
	if err != nil {
		return result, NewNightscoutError(err, "cant retreive list glucose entries")
	}

	return
}

type SVG int

func (svg SVG) HighOutOfRange(max int) string {
	if int(svg) >= max {
		return "true"
	} else {
		return "false"
	}
}

func (svg SVG) Float64() float64 {
	return float64(svg)
}

func (svg SVG) MMol() float64 {
	return svg.Float64() / 18
}

func (svg SVG) LowOutOfRange(min int) string {
	if int(svg) <= min {
		return "true"
	} else {
		return "false"
	}
}

type NSTime time.Time

func (t *NSTime) Time() time.Time {
	return time.Time(*t)
}

func (t *NSTime) UnmarshalJSON(data []byte) error {
	millis, err := strconv.ParseInt(string(data), 10, 64)
	if err != nil {
		return err
	}
	*t = NSTime(time.UnixMilli(millis))
	return nil
}

func (t *NSTime) MarshalJSON() ([]byte, error) {
	return []byte(strconv.FormatInt(t.Time().UnixMilli(), 10)), nil
}

func (t *NSTime) MarshalYAML() (interface{}, error) {
	return t.Time().UnixMilli(), nil
}

type GlucoseEntry struct {
	ID           string    `json:"_id"`
	App          string    `json:"app"`
	Device       string    `json:"device"`
	Date         *NSTime   `json:"date"`
	CreatedAt    string    `json:"created_at"`
	DateString   time.Time `json:"dateString"`
	Sgv          SVG       `json:"sgv"`
	Delta        float64   `json:"delta"`
	Direction    string    `json:"direction"`
	Type         string    `json:"type"`
	Filtered     float64   `json:"filtered"`
	Unfiltered   float64   `json:"unfiltered"`
	Units        string    `json:"units"`
	Rssi         int       `json:"rssi"`
	Noise        int       `json:"noise"`
	SysTime      time.Time `json:"sysTime"`
	DataType     int       `json:"dataType"`
	RecordNumber int       `json:"recordNumber"`
	UtcOffset    int       `json:"utcOffset"`
	Subject      string    `json:"subject"`
	Mills        int64     `json:"mills"`
}

func (g *GlucoseEntry) Kind() string {
	return "GlucoseEntry"
}

type GlucoseEntries []*GlucoseEntry

func (r *GlucoseEntries) Append(e *GlucoseEntry) {
	*r = append(*r, e)
}

func (r GlucoseEntries) Len() int {
	return len(r)
}

type GlucoseFilterFunc func(*GlucoseEntry) bool

func OnlyAfter(date time.Time) GlucoseFilterFunc {
	return func(e *GlucoseEntry) bool {
		return e.Date.Time().After(date)
	}
}

func (es *GlucoseEntries) Filter(fn GlucoseFilterFunc) (result *GlucoseEntries) {
	result = &GlucoseEntries{}
	es.Visit(func(e *GlucoseEntry, _ error) error {
		if fn(e) {
			result.Append(e)
		}
		return nil
	})
	return result
}

type DownsampleFunc func() time.Duration

func DownsampleDuration(d time.Duration) DownsampleFunc {
	return func() time.Duration {
		return d
	}
}

func (es *GlucoseEntries) Downsample(f DownsampleFunc) *GlucoseEntries {

	var lastTS *NSTime

	return es.Filter(func(e *GlucoseEntry) bool {
		if lastTS == nil {
			lastTS = e.Date
			return true
		}

		if lastTS.Time().Sub(e.Date.Time()) > f() {
			lastTS = e.Date
			return true
		}

		return false
	})

}

type VisitorFunc func(*GlucoseEntry, error) error

func (es GlucoseEntries) Visit(fn VisitorFunc) error {
	var err error
	for _, entry := range es {
		if err = fn(entry, err); err != nil {
			return err
		}
	}
	return nil
}
