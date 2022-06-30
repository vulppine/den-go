package routing

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type testRoute struct {
	// intended endpoint for this route
	endpoint string
	// intended method for response/request information
	method string
	// intended responseType to be sent
	responseType ResponseType
	// body that will be sent in the ResponseInfo
	body string
}

func (t *testRoute) ProcessResponse(resp *ResponseInfo) error {
	return nil
}

func (t *testRoute) ProcessRequest(req *http.Request) error {
	if req.Method != t.method {
		return fmt.Errorf("method of request did not match intended method: expected %s, got %s", t.method, req.Method)
	}

	return nil
}

func (t *testRoute) HandleRequest(req *RequestInfo) (*ResponseInfo, error) {
	if req.RequestEndpoint != t.endpoint {
		return nil, fmt.Errorf("endpoint did not match route endpoint: expected %s, got %s", t.endpoint, req.RequestEndpoint)
	}

	if req.Method() != t.method {
		return nil, fmt.Errorf("method of request did not match intended method: expected %s, got %s", t.method, req.Method())
	}

	return &ResponseInfo{
		code:         http.StatusOK,
		Headers:      nil,
		responseType: t.responseType,
		endpoint:     t.endpoint,
		Body:         bytes.NewBufferString(t.body),
	}, nil
}

type alwaysError struct {
	onResponse bool
	onRoute    bool
	onRequest  bool
}

func (a *alwaysError) ProcessResponse(*ResponseInfo) error {
	if a.onResponse {
		return errors.New("error on response")
	}

	return nil
}

func (a *alwaysError) ProcessRequest(*http.Request) error {
	if a.onRequest {
		return errors.New("error on request")
	}

	return nil
}

func (a *alwaysError) HandleRequest(*RequestInfo) (*ResponseInfo, error) {
	if a.onRoute {
		return nil, errors.New("error on route")
	}

	return &ResponseInfo{
		code:         http.StatusOK,
		Headers:      nil,
		responseType: None,
		endpoint:     "",
		Body:         nil,
	}, nil
}

func TestRouter_RegisterRoute(t *testing.T) {
	router := new(Router)
	handler := new(testRoute)
	req := new(RequestInfo)

	router.RegisterRoute("test", handler)

	routeHandler, err := router.getRouteHandler("test")

	if err != nil {
		t.Fatalf("Failed to get route handler for endpoint test: %s", err)
	}

	info, _ := routeHandler.HandleRequest(req)

	if info.endpoint != "test" && info.code != http.StatusOK {
		t.Fatalf("RegisterRoute(\"test\", handler): expected endpoint test, code 200, got %s, %d", info.endpoint, info.code)
	}
}

func TestRouter_RegisterRequestProcessor(t *testing.T) {
	router := new(Router)
	handler := new(testRoute)

	router.RegisterRequestProcessor(http.MethodGet, handler)

	handlers, err := router.getRequestProcessors(http.MethodGet)

	if err != nil {
		t.Fatalf("Failed to get request processors from handlers: %s", err)
	}

	if len(handlers) != 1 {
		t.Fatalf("Too many handlers, expected 1, got %d", len(handlers))
	}
}

func TestRouter_RegisterResponseProcessor(t *testing.T) {
	router := new(Router)
	handler := new(testRoute)

	router.RegisterResponseProcessor(None, "test", handler)

	handlers, err := router.getResponseProcessors(None, "test")

	if err != nil {
		t.Fatalf("Failed to get response processors from handlers: %s", err)
	}

	if len(handlers) != 1 {
		t.Fatalf("Too many handlers, expected 1, got %d", len(handlers))
	}
}

type dummyWriter struct {
	data    bytes.Buffer
	headers http.Header
	code    int
	final   []byte
}

func (d *dummyWriter) Write(i []byte) (int, error) {
	return d.data.Write(i)
}

func (d *dummyWriter) WriteHeader(statusCode int) {
	d.code = statusCode
	d.final = d.data.Bytes()
}

func (d *dummyWriter) Header() http.Header {
	return d.headers
}

func TestRouter_processRequest(t *testing.T) {
	router := new(Router)

	handler := new(testRoute)
	handler.method = http.MethodGet
	handler.endpoint = "endpoint"
	handler.responseType = Text
	handler.body = "Hello, world!"

	router.RegisterRoute(handler.endpoint, handler)
	router.RegisterResponseProcessor(handler.responseType, handler.endpoint, handler)

	ctx := newRoutingContext()
	testUrl, _ := url.Parse("https://test.org/endpoint/path/")
	req := &http.Request{
		Method: http.MethodGet,
		URL:    testUrl,
	}
	writer := new(dummyWriter)

	checkAdvance := func(stage routeStage) {
		select {
		case <-ctx.Done():
			t.Fatalf("context contains error: %s", ctx.Err())
		case s := <-ctx.stageChan:
			if s != stage {
				t.Fatalf("got different stage than expected: expected %d, got %d", stage, s)
			}

			ctx.stage = s
		}
	}

	router.processRequest(ctx, writer, req)
	checkAdvance(routing)

	router.processRequest(ctx, writer, req)

	if ctx.info == nil {
		t.Fatalf("expected info in context to be populated, got nil")
	}

	if ctx.info.ResponseType() != Text {
		t.Fatalf("expected Text response type")
	}

	checkAdvance(postProcess)

	router.processRequest(ctx, writer, req)

	if ctx.data == nil {
		t.Fatalf("expected data in context to be populated, got nil")
	}

	checkAdvance(send)

	router.processRequest(ctx, writer, req)
	checkAdvance(finish)

	if writer.code != http.StatusOK {
		t.Fatalf("expected code to be http.StatusOK, got %d", writer.code)
	}

	var s strings.Builder
	if _, err := s.Write(writer.final); err != nil {
		t.Fatalf("error writing to string builder: %s", err)
	}

	if s.String() != "Hello, world!" {
		t.Fatalf("expected Hello, world! in final body, got %s", s.String())
	}
}

func TestRouter_processRequestWithErrors(t *testing.T) {
	router := new(Router)

	handler := new(alwaysError)
	handler.onRequest = true

	router.RegisterRequestProcessor(http.MethodGet, handler)

	ctx := newRoutingContext()
	req := new(http.Request)
	req.Method = http.MethodGet
	writer := new(dummyWriter)

	advance := func() {
		select {
		case <-ctx.Done():
			return
		case <-ctx.stageChan:
			t.Fatalf("expected error, got success")
		}
	}

	router.processRequest(ctx, writer, req)
	advance()
}
