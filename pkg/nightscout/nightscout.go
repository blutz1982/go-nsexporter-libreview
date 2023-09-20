package nightscout

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	DefaultMaxSVG = 400
	DefaultMinSVG = 40
	MaxEnties     = 131072
)

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

type Treatment struct {
	ID        string    `json:"_id"`
	CreatedAt time.Time `json:"created_at"`
	Insulin   int       `json:"insulin"`
	Carbs     int       `json:"carbs"`
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

type GlucoseEntries []*GlucoseEntry

func (r *GlucoseEntries) Append(e *GlucoseEntry) {
	*r = append(*r, e)
}

func (r GlucoseEntries) Len() int {
	return len(r)
}

type FilterFunc func(*GlucoseEntry) bool

func OnlyAfter(date time.Time) FilterFunc {
	return func(e *GlucoseEntry) bool {
		return e.DateString.After(date)
	}
}

func (es GlucoseEntries) Filter(fn FilterFunc) (result GlucoseEntries) {
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

func (es GlucoseEntries) Downsample(f DownsampleFunc) GlucoseEntries {

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

type Client interface {
	GetGlucoseEntries(fromDate, toDate time.Time, count int) (entries GlucoseEntries, err error)
	GetGlucoseEntriesWithContext(ctx context.Context, fromDate, toDate time.Time, count int) (entries GlucoseEntries, err error)
	GetInsulinEntries(fromDate, toDate time.Time, count int) (entries Treatments, err error)
	GetInsulinEntriesWithContext(ctx context.Context, fromDate, toDate time.Time, count int) (entries Treatments, err error)
}

type nightscout struct {
	baseUrl  *url.URL
	apiToken string
	client   *http.Client
}

func NewWithConfig(config *Config) (Client, error) {

	u, err := url.Parse(config.URL)
	if err != nil {
		return nil, err
	}

	return &nightscout{
		baseUrl:  u,
		client:   http.DefaultClient,
		apiToken: config.APIToken,
	}, nil
}

func (ns *nightscout) GetInsulinEntries(fromDate, toDate time.Time, count int) (entries Treatments, err error) {
	return ns.GetInsulinEntriesWithContext(context.Background(), fromDate, toDate, count)
}

func (ns *nightscout) GetInsulinEntriesWithContext(ctx context.Context, fromDate, toDate time.Time, count int) (entries Treatments, err error) {

	url := ns.baseUrl.JoinPath("api", "v1", "treatments.json")
	q := url.Query()
	q.Add("find[created_at][$gte]", fromDate.UTC().Format(time.RFC3339))
	q.Add("find[created_at][$lte]", toDate.UTC().Format(time.RFC3339))
	q.Add("find[insulin][$gt]", "0")
	q.Add("count", strconv.Itoa(count))
	q.Add("token", ns.apiToken)
	url.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return entries, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := ns.client.Do(req)
	if err != nil {
		return entries, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return entries, fmt.Errorf("GetInsulinEntries: bad status code %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return entries, err
	}

	return entries, nil

}

func (ns *nightscout) GetGlucoseEntries(fromDate, toDate time.Time, count int) (entries GlucoseEntries, err error) {
	return ns.GetGlucoseEntriesWithContext(context.Background(), fromDate, toDate, count)
}

func (ns *nightscout) GetGlucoseEntriesWithContext(ctx context.Context, fromDate, toDate time.Time, count int) (entries GlucoseEntries, err error) {

	url := ns.baseUrl.JoinPath("api", "v1", "entries.json")
	q := url.Query()
	q.Add("find[dateString][$gte]", fromDate.UTC().Format(time.RFC3339))
	q.Add("find[dateString][$lte]", toDate.UTC().Format(time.RFC3339))
	q.Add("count", strconv.Itoa(count))
	q.Add("token", ns.apiToken)
	url.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url.String(), nil)
	if err != nil {
		return entries, err
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := ns.client.Do(req)
	if err != nil {
		return entries, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return entries, fmt.Errorf("GetGlucoseEntries: bad status code %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return entries, err
	}

	return entries, nil

}
