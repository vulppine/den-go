package files

import (
	"bytes"
	"den/routing"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// currently getFileAtPath just fetches the direct file, returning either:
// - the file in question
// - an error if the path attempted to go past the base path of the handler
// - an error if there was an issue getting the file from its given path
func TestFileHandler_getFileAtPathRelativePathing(t *testing.T) {
	f := &FileHandler{
		basePath: "/do_not_access/",
	}

	_, err := f.getFileAtPath("../")

	var e fileHandlerError

	if errors.As(err, &e) {
		if e.code != notAllowed {
			t.Fatalf("getFileAtPath: did not disallow delving past base path")
		}
	} else {
		t.Fatalf("no error detected (bug?)")
	}
}

// TODO: this should be some kind of common testing thing
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

func TestFileHandler_HandleRequest(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test_file")
	file, err := os.Create(testFile)

	if err != nil {
		t.Fatalf("error upon creating file: %s", err)
	}

	_, err = file.WriteString("Hello, world!")

	if err != nil {
		t.Fatalf("error upon write: %s", err)
	}

	f := NewFileHandler(dir)

	router := routing.NewRouter()
	router.RegisterRoute("test", f)
	testUrl, _ := url.Parse("https://test.org/test/test_file/")

	req := &http.Request{
		Method: http.MethodGet,
		URL:    testUrl,
	}

	w := new(dummyWriter)

	router.RouteRequest(w, req)

	var s strings.Builder

	s.Write(w.final)

	if w.code != http.StatusOK || len(w.final) == 0 {
		t.Fatalf("expected http.StatusOK with filled buffer, got code %d and body %s", w.code, s.String())
	}

	if s.String() != "Hello, world!" {
		t.Fatalf("expected Hello, world! as body, got %s", s.String())
	}
}

// ugh, copy and paste test sucks: TODO actually have some facilities for
// common testing structs and stuff
func TestFileHandler_HandleRequestWithRelativePathing(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "test_file")
	file, err := os.Create(testFile)

	if err != nil {
		t.Fatalf("error upon creating file: %s", err)
	}

	_, err = file.WriteString("Hello, world!")

	if err != nil {
		t.Fatalf("error upon write: %s", err)
	}

	f := NewFileHandler(dir)

	router := routing.NewRouter()
	router.RegisterRoute("test", f)
	testUrl, _ := url.Parse("https://test.org/test/../../../")

	req := &http.Request{
		Method: http.MethodGet,
		URL:    testUrl,
	}

	w := new(dummyWriter)

	router.RouteRequest(w, req)

	var s strings.Builder

	s.Write(w.final)

	if w.code != http.StatusForbidden || len(w.final) == 0 {
		t.Fatalf("expected http.StatusForbidden with filled buffer, got code %d and body %s", w.code, s.String())
	}
}
