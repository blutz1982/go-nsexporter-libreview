package nightscout

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/blutz1982/go-nsexporter-libreview/pkg/rest"
)

const (
	// Kind
	Insulin = "insulin"
	Carbs   = "carbs"
	Sgv     = "sgv"

	DefaultMaxSVG      = 400
	DefaultMinSVG      = 40
	MaxEnties          = 131072
	versionedAPIPathV1 = "api/v1"
	versionedAPIPathV2 = "api/v2"
)

var contentConfig = rest.ClientContentConfig{
	AcceptContentTypes: "application/json",
	ContentType:        "application/json",
}

type ListOptions struct {
	Kind     string
	DateFrom time.Time
	DateTo   time.Time
	Count    int
}

type Client interface {
	RESTClient() rest.Interface
	GlucoseGetter
	TreatmentsGetter
	DeviceGetter
}

type nightscout struct {
	restClient rest.Interface
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

	err = rest.NewRESTClient(u, versionedAPIPathV2, contentConfig, http.DefaultClient).
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

	client := http.DefaultClient
	client.Transport = rest.NewBearerAuthRoundTripper(JWTToken, http.DefaultTransport.(*http.Transport).Clone())

	return &nightscout{
		restClient: rest.NewRESTClient(u, versionedAPIPathV1, contentConfig, client),
	}, nil
}

func New(baseUrl string) (Client, error) {

	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	return &nightscout{
		restClient: rest.NewRESTClient(u, versionedAPIPathV1, contentConfig, http.DefaultClient),
	}, nil
}

func (ns *nightscout) RESTClient() rest.Interface {
	if ns == nil {
		return nil
	}
	return ns.restClient
}

func (ns *nightscout) Treatments() TreatmentInterface {
	return newTreatments(ns)
}

func (ns *nightscout) Glucose() GlucoseInterface {
	return newGlucose(ns)
}

func (ns *nightscout) DeviceStatus() DeviceInterface {
	return newDevice(ns)
}
