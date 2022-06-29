package routing

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

// TODO: Figure out how to do readonly fields in golang?

// RequestInfo is the information passed out of a Router into a RouteHandler.
// If creating this (which is not advised), it is not usable as a zero value, and requires use of
// NewRequestInfo, which is usually created upon a Router directing the request
// into a RouteHandler.
type RequestInfo struct {
	// Raw request for this struct. Expose fields as necessary.
	request *http.Request

	// Request endpoint that this request is fetching information from. This is either
	// from the subdomain, or from the first section of the path given in the request's URI.
	RequestEndpoint string

	Path []string // should be its own API because golang doesn't have a neat path thing but oh well?

	// Query of the request, from the URL.
	Query url.Values
}

// NewRequestInfo creates a new RequestInfo based on the request passed into it.
func NewRequestInfo(req *http.Request) *RequestInfo {
	info := RequestInfo{
		request: req,
	}

	info.getInfoFromUrl(req.URL)

	return &info
}

// Method exposes the HTTP request method to the caller.
func (i *RequestInfo) Method() string {
	return i.request.Method
}

// Body exposes the HTTP request's body to the caller. It is up to the caller
// if this is a valid body for the request or not.
func (i *RequestInfo) Body() io.ReadCloser {
	return i.request.Body
}

// Headers exposes the HTTP headers of this request to the caller.
func (i *RequestInfo) Headers() http.Header {
	return i.request.Header
}

// getInfoFromUrl gets the endpoint and the path from this URL, and fills in
// the respective fields for the RequestInfo struct given. Since URLs are complex,
// this only fetches the first section of the subdomain from the hostname, and
// if that isn't a valid subdomain (to be implemented: blacklist), it will instead
// use the first section of the path given to it.
func (i *RequestInfo) getInfoFromUrl(rawUrl *url.URL) {
	h := strings.Split(rawUrl.Hostname(), ".")
	p := strings.Split(strings.Trim(rawUrl.EscapedPath(), "/"), "/")

	if len(h) <= 2 {
		i.RequestEndpoint = p[0]
		i.Path = p[1:]
	} else {
		i.RequestEndpoint = h[0]
		i.Path = p
	}

	i.Query = rawUrl.Query()
}
