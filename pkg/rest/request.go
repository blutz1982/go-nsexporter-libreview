package rest

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"
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

	if len(c.content.ContentType) > 0 {
		r.SetHeader("Content-Type", c.content.ContentType)
	}

	if len(c.content.AcceptContentTypes) > 0 {
		r.SetHeader("Accept", c.content.AcceptContentTypes)
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

	serializer Serializer
}

type ResultError struct {
	err  error
	body []byte
}

func (r Result) Error() error {
	return ResultError{
		err:  r.err,
		body: r.body,
	}
}

func (resErr ResultError) Error() string {
	return fmt.Sprintf("%v: body : %s", resErr.err, resErr.body)
}

func (r Result) Into(obj any) error {
	if r.err != nil {
		return r.Error()
	}

	if len(r.body) == 0 {
		return fmt.Errorf("0-length response with status code: %d and content type: %s",
			r.statusCode, r.contentType)
	}

	if r.serializer == nil {
		return errors.New("no serializer")
	}

	return r.serializer.Decode(r.body, obj)
}

func (r *Request) transformResponse(resp *http.Response, req *http.Request) Result {

	// data, _ := httputil.DumpResponse(resp, true)
	// fmt.Println(string(data))

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
		contentType = r.c.content.ContentType
	}

	var serializer Serializer
	if len(contentType) > 0 {
		var err error
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			return Result{err: err}
		}

		var found bool
		serializer, found = SerializerForMediaType(mediaType)
		if !found {
			return Result{
				body: body,
				err:  errors.Errorf("no serializer found for mediatype %s", mediaType),
			}
		}

	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode > http.StatusPartialContent {
		return Result{
			body:        body,
			contentType: contentType,
			statusCode:  resp.StatusCode,
			err:         fmt.Errorf("bad status: %s at url %s", resp.Status, req.URL.String()),
		}
	}

	return Result{
		body:        body,
		contentType: contentType,
		statusCode:  resp.StatusCode,
		serializer:  serializer,
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

	// data, _ := httputil.DumpRequest(req, true)
	// fmt.Println(string(data))

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

func (r *Request) Body(obj any) *Request {
	if r.err != nil {
		return r
	}

	switch t := obj.(type) {
	case string:
		data, err := os.ReadFile(t)
		if err != nil {
			r.err = err
			return r
		}
		r.body = nil
		r.bodyBytes = data
	case []byte:
		r.body = nil
		r.bodyBytes = t
	case io.Reader:
		r.body = t
		r.bodyBytes = nil
	case Object:
		// callers may pass typed interface pointers, therefore we must check nil with reflection
		if reflect.ValueOf(t).IsNil() {
			return r
		}

		serializer, found := SerializerForMediaType(r.c.content.ContentType)
		if !found {
			r.err = fmt.Errorf("cant find serializer for ContentType: %s", r.c.content.ContentType)
			return r
		}

		buff := new(bytes.Buffer)
		err := serializer.Encode(t, buff)
		if err != nil {
			r.err = err
			return r
		}

		r.body = buff
		r.bodyBytes = nil
		r.SetHeader("Content-Type", r.c.content.ContentType)
	default:
		r.err = fmt.Errorf("unknown type used for body: %+v", obj)
	}
	return r
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
