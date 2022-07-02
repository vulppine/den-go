package pages

import "io"

type pageTree struct {
	root pageNode
}

type pageNode struct {
	// id of this node
	id string
	// children in this node as a map keyed by string, with values of itself:
	// if this node has a handler, this will be nil
	children map[string]pageNode
	// the handler of this node: if this node has children, this is always nil
	// adding children will always turn this into a nil value, but adding a handler
	// will never work if the node has children
	handler PageNodeHandler
}

func (n *pageNode) Handler() PageNodeHandler {
	return n.handler
}

// Child gets a child from the given page node.
func (n *pageNode) Child(key string) *pageNode {
	if n.children == nil {
		return nil
	}

	if node, ok := n.children[key]; ok {
		return &node
	}

	return nil
}

// Add adds a new child to the given page node.
func (n *pageNode) Add(key string) {
	if n.handler != nil {
		n.handler = nil
		n.children = make(map[string]pageNode)
	}

	n.children[key] = pageNode{id: key}
}

// SetHandler sets the handler to the given handler.
func (n *pageNode) SetHandler(handler PageNodeHandler) {
	if n.children != nil {
		return
	}

	n.handler = handler
}

// PageNodeHandler is an interface that is used when the node of a page tree
// is a leaf, and the node is accessed through searching through a page tree.
// This can be implemented by others, and while pageTree itself is a private
// struct, the APIs to access it from other Go packages are not.
type PageNodeHandler interface {
	// Page should always return a reader, because a leaf node of a page
	// tree *is* a page: therefore, something must be readable from
	// this page at least.
	Page(path []string) (io.Reader, error)
	// AllPages should return every page accessible from this node.
	// It should return a set of relative paths.
	AllPages() ([]string, error)
}
