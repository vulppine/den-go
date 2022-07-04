package pages

import (
	"bytes"
	"io"
	"strconv"
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

	a := tree.root.add("a")
	b := a.add("b")
	c := tree.root.add("c")

	handlerB := new(dummyPageNodeHandler)
	handlerB.toReturn = "b"

	handlerC := new(dummyPageNodeHandler)
	handlerC.toReturn = "c"

	b.setHandler(handlerB)
	c.setHandler(handlerC)

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

func TestPageTreeAddPath(t *testing.T) {
	tree := new(pageTree)

	pathA := []string{"x", "y", "z", "a"}
	pathB := []string{"x", "y", "b"}
	pathC := []string{"1", "2", "c"}

	handlers := make([]PageNodeHandler, 0)

	for i := 0; i < 3; i++ {
		handler := new(dummyPageNodeHandler)
		handler.toReturn = strconv.Itoa(i)
		handlers = append(handlers, handler)
	}

	tree.addPath(pathA, handlers[0])
	tree.addPath(pathB, handlers[1])
	tree.addPath(pathC, handlers[2])

	testValues := []struct {
		path  []string
		value string
	}{
		{pathA, "0"},
		{pathB, "1"},
		{pathC, "2"},
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
