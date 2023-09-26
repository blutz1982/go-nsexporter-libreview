package nightscout

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Request struct {
	c *RESTClient

	timeout time.Duration

	verb       string
	pathPrefix string
	params     url.Values
	headers    http.Header

	resourceName string
	resource     string

	// output
	err error

	// only one of body / bodyBytes may be set. requests using body are not retriable.
	body      io.Reader
	bodyBytes []byte
}

func NewRequest(c *RESTClient) *Request {

	var pathPrefix string
	if c.base != nil {
		pathPrefix = path.Join("/", c.base.Path, c.versionedAPIPath)
	} else {
		pathPrefix = path.Join("/", c.versionedAPIPath)
	}

	var timeout time.Duration
	if c.Client != nil {
		timeout = c.Client.Timeout
	}

	r := &Request{
		c:          c,
		timeout:    timeout,
		pathPrefix: pathPrefix,
	}

	if len(c.token) > 0 {
		r.SetHeader("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	if len(c.contentType) > 0 {
		r.SetHeader("Content-Type", c.contentType)
		r.SetHeader("Accept", c.contentType)
	}

	return r
}

func (r *Request) Verb(verb string) *Request {
	r.verb = verb
	return r
}

func (r *Request) Resource(resource string) *Request {
	if r.err != nil {
		return r
	}
	if len(r.resource) != 0 {
		r.err = fmt.Errorf("resource already set to %q, cannot change to %q", r.resource, resource)
		return r
	}

	r.resource = resource
	return r
}

func (r *Request) Name(resourceName string) *Request {
	if r.err != nil {
		return r
	}
	if len(resourceName) == 0 {
		r.err = fmt.Errorf("resource name may not be empty")
		return r
	}
	if len(r.resourceName) != 0 {
		r.err = fmt.Errorf("resource name already set to %q, cannot change to %q", r.resourceName, resourceName)
		return r
	}
	if msgs := IsValidPathSegmentName(resourceName); len(msgs) != 0 {
		r.err = fmt.Errorf("invalid resource name %q: %v", resourceName, msgs)
		return r
	}
	r.resourceName = resourceName
	return r
}

func (r *Request) setParam(paramName, value string) *Request {
	if r.params == nil {
		r.params = make(url.Values)
	}
	r.params[paramName] = append(r.params[paramName], value)
	return r
}

func (r *Request) Param(paramName, s string) *Request {
	if r.err != nil {
		return r
	}
	return r.setParam(paramName, s)
}

func (r *Request) SetHeader(key string, values ...string) *Request {
	if r.headers == nil {
		r.headers = http.Header{}
	}
	r.headers.Del(key)
	for _, value := range values {
		r.headers.Add(key, value)
	}
	return r
}

type Result struct {
	body        []byte
	contentType string
	err         error
	statusCode  int
}

func (r Result) Error() error {
	return r.err
}

func (r Result) Into(obj any) error {
	if r.err != nil {
		return r.Error()
	}

	if len(r.body) == 0 {
		return fmt.Errorf("0-length response with status code: %d and content type: %s",
			r.statusCode, r.contentType)
	}

	return json.Unmarshal(r.body, obj)
}

func (r *Request) transformResponse(resp *http.Response, req *http.Request) Result {

	var body []byte
	if resp.Body != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return Result{
				err: err,
			}
		}
		body = data
	}

	contentType := resp.Header.Get("Content-Type")
	if len(contentType) == 0 {
		contentType = r.c.contentType
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusPartialContent {
		return Result{
			body:        body,
			contentType: contentType,
			statusCode:  resp.StatusCode,
			err:         fmt.Errorf("bad status code %d", resp.StatusCode),
		}
	}

	return Result{
		body:        body,
		contentType: contentType,
		statusCode:  resp.StatusCode,
	}
}

func (r *Request) Do(ctx context.Context) Result {
	var result Result
	err := r.request(ctx, func(req *http.Request, resp *http.Response) {
		result = r.transformResponse(resp, req)
	})
	if err != nil {
		return Result{err: err}
	}

	return result
}

func (r *Request) request(ctx context.Context, fn func(*http.Request, *http.Response)) error {
	client := r.c.Client
	if client == nil {
		client = http.DefaultClient
	}

	req, err := r.newHTTPRequest(ctx)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	fn(req, resp)

	return nil
}

func (r *Request) newHTTPRequest(ctx context.Context) (*http.Request, error) {
	var body io.Reader
	switch {
	case r.body != nil && r.bodyBytes != nil:
		return nil, fmt.Errorf("cannot set both body and bodyBytes")
	case r.body != nil:
		body = r.body
	case r.bodyBytes != nil:
		body = bytes.NewReader(r.bodyBytes)
	}

	url := r.URL().String()
	req, err := http.NewRequest(r.verb, url, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header = r.headers
	return req, nil
}

// URL returns the current working URL.
func (r *Request) URL() *url.URL {
	p := r.pathPrefix

	if len(r.resource) != 0 {
		p = path.Join(p, strings.ToLower(r.resource))
	}

	if len(r.resourceName) != 0 {
		p = path.Join(p, r.resourceName)
	}

	finalURL := &url.URL{}
	if r.c.base != nil {
		*finalURL = *r.c.base
	}
	finalURL.Path = p

	query := url.Values{}
	for key, values := range r.params {
		for _, value := range values {
			query.Add(key, value)
		}
	}

	finalURL.RawQuery = query.Encode()
	return finalURL
}

var NameMayNotBe = []string{".", ".."}
var NameMayNotContain = []string{"/", "%"}

func IsValidPathSegmentName(name string) []string {
	for _, illegalName := range NameMayNotBe {
		if name == illegalName {
			return []string{fmt.Sprintf(`may not be '%s'`, illegalName)}
		}
	}

	var errors []string
	for _, illegalContent := range NameMayNotContain {
		if strings.Contains(name, illegalContent) {
			errors = append(errors, fmt.Sprintf(`may not contain '%s'`, illegalContent))
		}
	}

	return errors
}
