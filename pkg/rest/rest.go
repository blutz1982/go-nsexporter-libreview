package rest

import (
	"net/http"
	"net/url"
	"strings"
)

type Interface interface {
	Verb(verb string) *Request
	Get() *Request
	Post() *Request
	Delete() *Request
}

type ClientContentConfig struct {
	AcceptContentTypes string
	ContentType        string
}

type RESTClientOpt func(*RESTClient)

type RESTClient struct {
	base             *url.URL
	Client           *http.Client
	content          ClientContentConfig
	versionedAPIPath string
}

func NewRESTClient(baseURL *url.URL, versionedAPIPath string, config ClientContentConfig, client *http.Client, opts ...RESTClientOpt) *RESTClient {

	if len(config.ContentType) == 0 {
		config.ContentType = "application/json"
	}

	base := *baseURL
	if !strings.HasSuffix(base.Path, "/") {
		base.Path += "/"
	}
	base.RawQuery = ""
	base.Fragment = ""

	c := &RESTClient{
		base:             &base,
		content:          config,
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
	return c.Verb(http.MethodGet)
}

func (c *RESTClient) Post() *Request {
	return c.Verb(http.MethodPost)
}

func (c *RESTClient) Delete() *Request {
	return c.Verb(http.MethodDelete)
}
