package rest

import (
	"fmt"
	"net/http"
)

type bearerAuthRoundTripper struct {
	bearer string
	next   http.RoundTripper
}

func (rt *bearerAuthRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("Authorization")) != 0 || len(rt.bearer) == 0 {
		return rt.next.RoundTrip(req)
	}

	req = req.Clone(req.Context())

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", rt.bearer))
	return rt.next.RoundTrip(req)

}

func NewBearerAuthRoundTripper(bearer string, rt http.RoundTripper) http.RoundTripper {
	return &bearerAuthRoundTripper{bearer, rt}
}
