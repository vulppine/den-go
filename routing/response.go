package routing

import (
	"bytes"
	"io"
	"net/http"
)

// ResponseInfo contains the intended response code, the type of response that is
// being sent, the origin endpoint of the response, and the body of the response
// as raw bytes. This can be processed multiple times in a single request,
// dependent on the response type itself. After processing, this will be
// converted into ResponseData.
type ResponseInfo struct {
	// code represents the HTTP status code that the response will give. It is
	// readonly, as if an error occurs midway through the pipeline, the response
	// should be discarded, and replaced with the error response.
	code int

	// Headers represents the HTTP headers that will be sent along
	// with the response. This is public, as HTTP headers can be
	// switched out if necessary during the post process.
	Headers http.Header

	// responseType is the response type that this response represents.
	// It is readonly, to prevent any type changes during post process.
	// A post process function can change the body instead for the end
	// user, no matter what the response is.
	responseType ResponseType

	// endpoint is the endpoint that this response originated from.
	// It is readonly, to prevent modification during the post
	// process stage. It is also mostly useless to read from, as
	// the router should handle this.
	endpoint string

	// Body is the body of the response. It can be freely changed in
	// post process in order to give this response a different body
	// after the request has been processed. This is a reader, to avoid
	// doing something like loading an entire data file into memory if
	// the responseType is of type Data.
	Body io.Reader
}

// CreateResponseInfo creates a new ResponseInfo for use with the rest of the pipeline.
// Since ResponseInfo has private fields (to avoid mutating any important detail fields),
// this is required.
func CreateResponseInfo(code int, headers http.Header, responseType ResponseType, endpoint string, body io.Reader) ResponseInfo {
	return ResponseInfo{
		code,
		headers,
		responseType,
		endpoint,
		body,
	}
}

// CreateGenericErrorResponse creates a generic error response using the given
// HTTP code and text message. The resultant ResponseInfo is a text response with
// the reserved endpoint, EndpointError.
func CreateGenericErrorResponse(code int, msg string) *ResponseInfo {
	return &ResponseInfo{
		code:         code,
		responseType: Text,
		endpoint:     EndpointError,
		Body:         bytes.NewBufferString(msg),
	}
}

func (i *ResponseInfo) ResponseType() ResponseType {
	return i.responseType
}

func (i *ResponseInfo) Code() int {
	return i.code
}

func (i *ResponseInfo) Finalize() ResponseData {
	return ResponseData{
		Code:    i.code,
		Headers: i.Headers,
		Data:    i.Body,
	}
}

// ResponseData is the final leg of the pipeline, where it only contains
// headers, the response code, and the data to send to the client. This
// should not be used for anything aside from sending directly to the
// client, and may be privatized in the future once the HTTP library is
// fully organized.
type ResponseData struct {
	Code    int
	Headers http.Header
	Data    io.Reader
}

// Send sends the response data over the given ResponseWriter. If an
// error occurs during writing, it will write a http.StatusServiceUnavailable
// into the response and then immediately stop writing any more data.
func (data *ResponseData) send(w http.ResponseWriter) {
	headers := w.Header()

	// add every single header into the set of headers
	// unless the header was set to nil at some point,
	// then just remove the header entirely and continue
	for k, v := range data.Headers {
		if v == nil {
			headers.Del(k)
			continue
		}

		for _, h := range v {
			headers.Add(k, h)
		}
	}

	buf := make([]byte, ChunkSize)

	var b int
	var readErr error
	var writeErr error

	if data.Data != nil {
		for readErr != io.EOF {
			if b, readErr = data.Data.Read(buf); readErr != nil && readErr != io.EOF {
				// error handle
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}

			if _, writeErr = w.Write(buf[0:b]); writeErr != nil {
				// error handle
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
		}
	}

	w.WriteHeader(data.Code)
}

type ResponseType int

const (
	// Html type. Use this if you're sending
	// raw HTML in your response.
	Html ResponseType = iota
	// Text type. Use this if you're sending only
	// text in your response.
	Text
	// Json type. Use this if you're sending only
	// JSON in your response.
	Json
	// Data type. Use this if you're sending raw
	// data through the stream.
	Data
	// None type. Use this if you're sending no data
	// through the response.
	None
)
