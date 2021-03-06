package pages

import "io"

type pageTree struct {
	root pageNode
}

func (t *pageTree) getHandler(path []string) (PageNodeHandler, []string) {
	n := &t.root

	for n.hasChildren() {
		c := n.child(path[0])
		if c == nil {
			return nil, path
		}

		n = c
		path = path[1:]
	}

	return n.handler, path
}

func (t *pageTree) addPath(path []string, handler PageNodeHandler) {
	if len(path) == 0 {
		return
	}

	n := &t.root

	for len(path) > 0 {
		if v, ok := n.children[path[0]]; !ok {
			n = n.add(path[0])
		} else {
			n = v
		}

		path = path[1:]
	}

	n.handler = handler
}

type pageNode struct {
	// id of this node
	id string
	// children in this node as a map keyed by string, with values of itself:
	// if this node has a handler, this will be nil
	children map[string]*pageNode
	// the handler of this node: if this node has children, this is always nil
	// adding children will always turn this into a nil value, but adding a handler
	// will never work if the node has children
	handler PageNodeHandler
}

func (n *pageNode) hasChildren() bool {
	return n.children != nil
}

// child gets a child from the given page node.
func (n *pageNode) child(key string) *pageNode {
	if n.children == nil {
		return nil
	}

	if node, ok := n.children[key]; ok {
		return node
	}

	return nil
}

// allChildren gets all children from the node into a slice of pageNode pointers.
func (n *pageNode) allChildren() []*pageNode {
	res := make([]*pageNode, 1)

	for _, v := range n.children {
		res = append(res, v)
	}

	return res
}

// add will add a new child to the given page node.
func (n *pageNode) add(key string) *pageNode {
	if n.handler != nil {
		n.handler = nil
	}

	if n.children == nil {
		n.children = make(map[string]*pageNode)
	}

	node := &pageNode{id: key}
	n.children[key] = node

	return node
}

// setHandler sets the handler to the given handler.
func (n *pageNode) setHandler(handler PageNodeHandler) {
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
