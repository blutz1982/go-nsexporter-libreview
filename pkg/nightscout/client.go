package nightscout

import (
	"net/http"
	"net/url"
	"time"
)

const (
	Insulin       = "insulin"
	Carbs         = "carbs"
	DefaultMaxSVG = 400
	DefaultMinSVG = 40
	MaxEnties     = 131072
)

type GetOptions struct {
	APIToken string
	DateFrom time.Time
	DateTo   time.Time
	Count    int
}

type Client interface {
	RESTClient() RestInterface
	GlucoseGetter
	TreatmentsGetter
}

type nightscout struct {
	restClient RestInterface
	// apiToken   string
}

func New(host string) (Client, error) {

	u, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	return &nightscout{
		restClient: NewRESTClient(u, "api/v1", http.DefaultClient),
	}, nil
}

func (ns *nightscout) RESTClient() RestInterface {
	if ns == nil {
		return nil
	}
	return ns.restClient
}

func (ns *nightscout) Treatments(kind string) TreatmentInterface {
	return newTreatments(ns, kind)
}

func (ns *nightscout) Glucose() GlucoseInterface {
	return newGlucose(ns)
}
