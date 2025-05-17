package aiclient

import "net/http"

type tripperware func(req *http.Request) (*http.Response, error)

func (t tripperware) RoundTrip(req *http.Request) (*http.Response, error) {
	return t(req)
}
