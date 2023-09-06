package nightscout

import (
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

	TimestampLayout = "2006-01-02"
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

type GlucoseEntries []*GlucoseEntry

func (r *GlucoseEntries) Append(e *GlucoseEntry) {
	*r = append(*r, e)
}

func (r GlucoseEntries) Len() int {
	return len(r)
}

func (es GlucoseEntries) Filter(fn func(*GlucoseEntry) bool) (result GlucoseEntries) {
	es.Visit(func(e *GlucoseEntry, _ error) error {
		if fn(e) {
			result.Append(e)
		}
		return nil
	})
	return result
}

type DownsampleFunc func() float64

func DownsampleMinutes(minutes int) DownsampleFunc {
	return func() float64 {
		return float64(minutes)
	}
}

// func (es GlucoseEntries) Downsample(minutes int) GlucoseEntries {
func (es GlucoseEntries) Downsample(f DownsampleFunc) GlucoseEntries {

	var lastTS *time.Time

	return es.Filter(func(e *GlucoseEntry) bool {
		if lastTS == nil {
			lastTS = &e.DateString
			return true
		}

		if lastTS.Sub(e.DateString).Minutes() > f() {
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

func (ns *nightscout) GetGlucoseEntries(fromDate, toDate time.Time, count int) (entries GlucoseEntries, err error) {

	url := ns.baseUrl.JoinPath("api", "v1", "entries.json")
	q := url.Query()
	q.Add("find[dateString][$gte]", fromDate.Format(TimestampLayout))
	q.Add("find[dateString][$lte]", toDate.Format(TimestampLayout))
	q.Add("count", strconv.Itoa(count))
	q.Add("token", ns.apiToken)
	url.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
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
