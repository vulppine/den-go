package routing

import (
	"errors"
	"fmt"
	"net/http"
)

// Router is the core of den, it is what links the request to the module.
// Before I fully write this documentation, there should be a reminder that
// errors should always propagate upwards so that state upgrades do not occur.
// This behavior is subject to change in the (near) future, as I may instead
// just let states upgrade until it reaches send (which is guaranteed not to
// error out the context, because any errors that occur there instead are just
// directly sent to the client)
type Router struct {
	// Map of routes in this router, by string. Each route
	// leads to a RouteHandler, an entry point into another
	// registered module.
	routes map[string]RouteHandler

	// Map of request processors to HTTP methods.
	requestProcessors map[string][]RequestProcessor

	// Map of processors to response types. This is intended
	// for anything that occurs after the data is processed
	// by the endpoint handler.
	responseProcessors map[ResponseType]responseProcessors
}

func NewRouter() *Router {
	router := new(Router)

	router.routes = make(map[string]RouteHandler)
	router.requestProcessors = make(map[string][]RequestProcessor)
	router.responseProcessors = make(map[ResponseType]responseProcessors)

	return router
}

// ChunkSize dictates how many bytes to write to the client
// while sending the final chunk of ResponseData. This should
// be user configurable in the future.
const ChunkSize = 1024 * 10

const (
	// EndpointRoot indicates that the endpoint reached is
	// the root of the website, or the empty string.
	EndpointRoot = ""
	// EndpointDefault indicates that the endpoint that was given
	// is not registered, so it will instead attempt to use the
	// default endpoint handler.
	EndpointDefault = "___DEFAULT___"

	// EndpointError indicates that an error occurred during data
	// processing internally, and that the resultant data will
	// immediately return a generic error. This may or may not
	// be privatized in the future to be replaced with generic
	// debug information to present to the user, processed by
	// a response processor, to avoid any panics due to
	// bad implementations of the ResponseProcessor.
	EndpointError = "___ERROR___"
)

// responseProcessors are a map of ResponseProcessor to endpoints.
type responseProcessors map[string][]ResponseProcessor

// RequestProcessor is an interface for processing requests before
// the request is routed to an endpoint.
type RequestProcessor interface {
	ProcessRequest(req *http.Request) error
}

type RouteHandler interface {
	HandleRequest(req *RequestInfo) (*ResponseInfo, error)
}

type ResponseProcessor interface {
	ProcessResponse(resp *ResponseInfo) error
}

// RegisterRoute registers an endpoint to a route handler.
func (r *Router) RegisterRoute(route string, handler RouteHandler) {
	if r.routes == nil {
		r.routes = make(map[string]RouteHandler)
	}

	r.routes[route] = handler
}

// RegisterRequestProcessor registers an HTTP method to a request processor.
// This is before the endpoint is parsed from the request, so this is mostly
// useful to fetch any important details from the HTTP request being made.
func (r *Router) RegisterRequestProcessor(method string, handler RequestProcessor) {
	if r.requestProcessors == nil {
		r.requestProcessors = make(map[string][]RequestProcessor)
	}

	r.requestProcessors[method] = append(r.requestProcessors[method], handler)
}

// RegisterResponseProcessor registers a response type, and an endpoint it originates from
// to a ResponseProcessor. This is useful for when you want to process a response and transform
// it based on the endpoint that the original request was attempting to access.
func (r *Router) RegisterResponseProcessor(responseType ResponseType, endpoint string, handler ResponseProcessor) {
	if r.responseProcessors == nil {
		r.responseProcessors = make(map[ResponseType]responseProcessors)
	}

	if r.responseProcessors[responseType] == nil {
		r.responseProcessors[responseType] = make(responseProcessors)
	}

	r.responseProcessors[responseType][endpoint] = append(r.responseProcessors[responseType][endpoint], handler)
}

func (r *Router) getRouteHandler(endpoint string) (RouteHandler, error) {
	if handler, ok := r.routes[endpoint]; ok {
		return handler, nil
	} else if handler, ok := r.routes[EndpointDefault]; ok {
		return handler, nil
	}

	return nil, errors.New("could not retrieve handler")
}

func (r *Router) getRequestProcessors(method string) ([]RequestProcessor, error) {
	if handlers, ok := r.requestProcessors[method]; ok {
		return handlers, nil
	} else {
		r.requestProcessors[method] = make([]RequestProcessor, 0)
		return r.requestProcessors[method], nil
	}
}

func (r *Router) getResponseProcessors(responseType ResponseType, endpoint string) ([]ResponseProcessor, error) {
	if typeHandlers, ok := r.responseProcessors[responseType]; ok {
		if endpointHandlers, ok := typeHandlers[endpoint]; ok {
			return endpointHandlers, nil
		}

		return nil, errors.New("no handlers registered for this endpoint")
	}

	return nil, errors.New("no handlers registered for this method")
}

func (r *Router) preProcessRequest(ctx *routingContext, req *http.Request) error {
	var err error
	if handler, err := r.getRequestProcessors(req.Method); err == nil {
		for _, h := range handler {
			err := h.ProcessRequest(req)
			if err != nil {
				ctx.CloseWithError(err)
			}
		}
	} else {
		ctx.CloseWithError(err)
		return err
	}

	return err
}

func (r *Router) handleRequest(ctx *routingContext, req *http.Request) (*ResponseInfo, error) {
	info := NewRequestInfo(req)

	if handler, err := r.getRouteHandler(info.requestEndpoint); err == nil {
		resp, err := handler.HandleRequest(info)

		if err != nil {
			ctx.CloseWithError(err)
		}

		return resp, err
	}

	// TODO error
	return nil, errors.New("could not find endpoint")
}

func (r *Router) processResponse(ctx *routingContext, resp *ResponseInfo) (*ResponseData, error) {
	if handlers, err := r.getResponseProcessors(resp.ResponseType(), resp.endpoint); err == nil {
		for _, h := range handlers {
			err := h.ProcessResponse(resp)
			if err != nil {
				ctx.CloseWithError(err)
				return nil, err
			}
		}
	}

	res := resp.Finalize()

	return &res, nil
}

// routeStage is an indicator of what stage the router is going through.
// This is essentially a linear state machine; this is so that Contexts
// can be cancelled in between requests
type routeStage int

const (
	initial routeStage = iota
	routing
	postProcess
	send
	finish
)

func (r *Router) processRequest(ctx *routingContext, w http.ResponseWriter, req *http.Request) {
	var err error
	switch ctx.stage {
	case initial:
		err = r.preProcessRequest(ctx, req)
	case routing:
		ctx.info, err = r.handleRequest(ctx, req)
	case postProcess:
		ctx.data, err = r.processResponse(ctx, ctx.info)
	case send:
		ctx.data.send(w)
	}

	// if there's no error, or the context already
	// has a failed state, upgrade the state

	// the reasoning behind context stage upgrades
	// with a failed state is to ensure that the
	// data is forced through all stages until it
	// reaches the client (as the contained response
	// should now be a generic error)
	if err == nil || ctx.Err() != nil {
		ctx.upgradeStage()
	}
}

// RouteRequest is a function that routes a request into the router tables.
// The request is routed through in a goroutine, and then blocked until either
// the context is cancelled (most likely due to an error), or until the 'finish'
// stage is reached in the routing state.
func (r *Router) RouteRequest(w http.ResponseWriter, req *http.Request) {
	ctx := newRoutingContext()

	// loop through until the route reaches the final state
	// even with an error, it will always reach the final state
	for ctx.stage != finish {
		go r.processRequest(ctx, w, req)

		select {
		case <-ctx.Done():
			if ctx.stage != finish {
				ctx.info = CreateGenericErrorResponse(http.StatusServiceUnavailable, fmt.Sprint(ctx.Err()))
				ctx.stage = postProcess
			}
		case s := <-ctx.stageChan:
			ctx.stage = s
		}
	}
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.RouteRequest(w, req)
}
