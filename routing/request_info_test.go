package routing

import (
	"net/http"
	"net/url"
	"testing"
)

func TestNewRequestInfoRootEndpoint(t *testing.T) {
	req := new(http.Request)
	req.URL, _ = url.Parse("https://test.org/")

	info := NewRequestInfo(req)

	if info.requestEndpoint != EndpointRoot {
		t.Errorf("expected EndpointRoot as endpoint, got %s", info.requestEndpoint)
	}
}

func TestNewRequestInfoSubdomainEndpoint(t *testing.T) {
	req := new(http.Request)
	req.URL, _ = url.Parse("https://endpoint.test.org/notendpoint/")

	info := NewRequestInfo(req)

	if info.requestEndpoint != "endpoint" {
		t.Errorf("expected endpoint as endpoint, got %s", info.requestEndpoint)
	}

	if info.Path[0] != "notendpoint" {
		t.Errorf("expected notendpoint as first section of path, got %s", info.Path[0])
	}
}
