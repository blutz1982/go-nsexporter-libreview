package nightscout

import (
	"context"
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
	URLToken string
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
	// urlToken   string
}

type TokenResponse struct {
	Token            string     `json:"token"`
	Sub              string     `json:"sub"`
	PermissionGroups [][]string `json:"permissionGroups"`
	Iat              int        `json:"iat"`
	Exp              int        `json:"exp"`
}

func NewJWTToken(baseUrl string, urlToken string) (string, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}

	tokenResp := new(TokenResponse)

	err = NewRESTClient(u, "/api/v2", http.DefaultClient).
		Get().
		Resource("authorization/request").
		Name(urlToken).
		Do(context.Background()).
		Into(tokenResp)
	if err != nil {
		return "", err
	}

	return tokenResp.Token, nil
}

func NewWithJWTToken(baseUrl, JWTToken string) (Client, error) {

	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	return &nightscout{
		restClient: NewRESTClient(u, "api/v1", http.DefaultClient, WithJWTToken(JWTToken)),
	}, nil
}

func New(baseUrl string) (Client, error) {

	u, err := url.Parse(baseUrl)
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
