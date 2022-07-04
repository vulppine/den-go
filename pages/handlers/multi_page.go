package handlers

import (
	"errors"
	"io"
	"strings"
)

// MultiPageNodeHandler is a map of readers to strings that correspond to the
// rest of the leftover path from the original tree traversal.
type MultiPageNodeHandler struct {
	pages map[string]io.Reader
}

func (m *MultiPageNodeHandler) Page(path []string) (io.Reader, error) {
	if v, ok := m.pages[strings.Join(path, "/")]; ok {
		return v, nil
	}

	return nil, errors.New("error fetching page")
}

func (m *MultiPageNodeHandler) AllPages() ([]string, error) {
	res := make([]string, 0)

	for k := range m.pages {
		res = append(res, k)
	}

	return res, nil
}
