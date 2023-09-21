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

type RESTClient struct {
	base             *url.URL
	Client           *http.Client
	contentType      string
	versionedAPIPath string
}

func NewRESTClient(baseURL *url.URL, versionedAPIPath string, client *http.Client) *RESTClient {

	base := *baseURL
	if !strings.HasSuffix(base.Path, "/") {
		base.Path += "/"
	}
	base.RawQuery = ""
	base.Fragment = ""

	return &RESTClient{
		base:             &base,
		contentType:      "application/json",
		versionedAPIPath: versionedAPIPath,
		Client:           client,
	}
}

func (c *RESTClient) Verb(verb string) *Request {
	return NewRequest(c).Verb(verb)
}

func (c *RESTClient) Get() *Request {
	return c.Verb("GET")
}
