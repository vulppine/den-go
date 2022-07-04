package handlers

import "io"

// SinglePageNodeHandler returns a single page. This can be any reader.
type SinglePageNodeHandler struct {
	name string
	body io.Reader
}

// NewSinglePageNodeHandler returns a SinglePageNodeHandler with the given name and body.
func NewSinglePageNodeHandler(name string, body io.Reader) *SinglePageNodeHandler {
	return &SinglePageNodeHandler{name, body}
}

func (s *SinglePageNodeHandler) Page([]string) (io.Reader, error) {
	return s.body, nil
}

func (s *SinglePageNodeHandler) AllPages() ([]string, error) {
	return []string{s.name}, nil
}
