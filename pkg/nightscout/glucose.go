package nightscout

import (
	"context"
	"strconv"
	"time"
)

type GlucoseGetter interface {
	Glucose() GlucoseInterface
}

type GlucoseInterface interface {
	Get(ctx context.Context, opts GetOptions) (*GlucoseEntries, error)
}

type glucose struct {
	client RestInterface
}

// newGlucose returns a Glucos
func newGlucose(c Client) *glucose {
	return &glucose{
		client: c.RESTClient(),
	}
}

func (g glucose) Get(ctx context.Context, opts GetOptions) (result *GlucoseEntries, err error) {
	result = &GlucoseEntries{}
	r := g.client.Get().
		Name("entries").
		Param("find[dateString][$gte]", opts.DateFrom.UTC().Format(time.RFC3339)).
		Param("find[dateString][$lte]", opts.DateTo.UTC().Format(time.RFC3339)).
		Param("count", strconv.Itoa(opts.Count))

	if len(opts.URLToken) > 0 {
		r = r.Param("token", opts.URLToken)
	}

	err = r.Do(ctx).Into(result)

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

func (svg SVG) LowOutOfRange(min int) string {
	if int(svg) <= min {
		return "true"
	} else {
		return "false"
	}
}

type GlucoseEntry struct {
	ID           string    `json:"_id"`
	Device       string    `json:"device"`
	Date         int64     `json:"date"`
	DateString   time.Time `json:"dateString"`
	Sgv          SVG       `json:"sgv"`
	Delta        float64   `json:"delta"`
	Direction    string    `json:"direction"`
	Type         string    `json:"type"`
	Filtered     float64   `json:"filtered"`
	Unfiltered   float64   `json:"unfiltered"`
	Rssi         int       `json:"rssi"`
	Noise        int       `json:"noise"`
	SysTime      time.Time `json:"sysTime"`
	DataType     int       `json:"dataType"`
	RecordNumber int       `json:"recordNumber"`
	UtcOffset    int       `json:"utcOffset"`
	Mills        int64     `json:"mills"`
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
		return e.DateString.After(date)
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

	var lastTS *time.Time

	return es.Filter(func(e *GlucoseEntry) bool {
		if lastTS == nil {
			lastTS = &e.DateString
			return true
		}

		if lastTS.Sub(e.DateString) > f() {
			lastTS = &e.DateString
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
