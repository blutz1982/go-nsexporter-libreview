package nightscout

import (
	"net/http"
	"net/url"
	"strings"
)

type RestInterface interface {
	Verb(verb string) *Request
	Get() *Request
}

type RESTClientOpt func(*RESTClient)

func WithJWTToken(token string) RESTClientOpt {
	return func(c *RESTClient) {
		c.token = token
	}
}

type RESTClient struct {
	base             *url.URL
	Client           *http.Client
	contentType      string
	versionedAPIPath string
	token            string
}

func NewRESTClient(baseURL *url.URL, versionedAPIPath string, client *http.Client, opts ...RESTClientOpt) *RESTClient {

	base := *baseURL
	if !strings.HasSuffix(base.Path, "/") {
		base.Path += "/"
	}
	base.RawQuery = ""
	base.Fragment = ""

	c := &RESTClient{
		base:             &base,
		contentType:      "application/json",
		versionedAPIPath: versionedAPIPath,
		Client:           client,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *RESTClient) Verb(verb string) *Request {
	return NewRequest(c).Verb(verb)
}

func (c *RESTClient) Get() *Request {
	return c.Verb("GET")
}
