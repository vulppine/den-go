package pages

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

type dummyPageNodeHandler struct {
	toReturn string
}

func (d *dummyPageNodeHandler) Page(path []string) (io.Reader, error) {
	return bytes.NewBufferString(d.toReturn), nil
}

func (d *dummyPageNodeHandler) AllPages() ([]string, error) {
	return []string{}, nil
}

func TestPageTreeTraversal(t *testing.T) {
	tree := new(pageTree)

	a := tree.root.Add("a")
	b := a.Add("b")
	c := tree.root.Add("c")

	handlerB := new(dummyPageNodeHandler)
	handlerB.toReturn = "b"

	handlerC := new(dummyPageNodeHandler)
	handlerC.toReturn = "c"

	b.SetHandler(handlerB)
	c.SetHandler(handlerC)

	testValues := []struct {
		path  []string
		value string
	}{
		{[]string{"a", "b"}, "b"},
		{[]string{"c"}, "c"},
	}

	for _, v := range testValues {
		handler, path := tree.getHandler(v.path)

		reader, _ := handler.Page(path)
		rawBytes, _ := io.ReadAll(reader)
		var s strings.Builder

		s.Write(rawBytes)

		if s.String() != v.value {
			t.Fatalf("expected string %s from handler, got: %s", v.value, s.String())
		}
	}

}
