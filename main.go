package fway

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
)

type Value string

func Params(r *http.Request) map[string]string {
	return r.Context().Value(Value("params")).(map[string]string)
}

type node struct {
	child     []*node
	wildChild []*node
	path      string
	part      string
	isWild    bool
	handler   http.HandlerFunc
}

func (n *node) String(i int) string {
	var b bytes.Buffer

	if n.path == "" {
		b.WriteString("|-> [ uri: '/' ]\n")
	} else {
		b.WriteString(strings.Repeat("  ", i) + "|-> ")
		part := n.part
		if n.isWild {
			part = ":" + n.part
		}

		b.WriteString(part)
		b.WriteString(fmt.Sprintf(" [ uri: '%s' handler: %v ]", n.path, !(n.handler == nil)))
		b.WriteString("\n")
	}

	for _, v := range n.child {
		b.WriteString(v.String(i + 1))
	}

	for _, v := range n.wildChild {
		b.WriteString(v.String(i + 1))
	}

	return b.String()
}

func (n *node) search(path string) (*node, map[string]string) {
	params := map[string]string{}
	parts := strings.Split(path, "/")

	var currNode = n
	for _, part := range parts {
		var foundNode *node = nil

		for _, child := range currNode.child {
			if child.part == part {
				foundNode = child
				break
			}
		}

		if foundNode == nil {
			// Now only take in account first wild node registered
			// TODO: add support for multiple wild nodes with regex
			for _, child := range currNode.wildChild {
				foundNode = child
				params[child.part] = part
				break
			}
		}

		if foundNode == nil {
			return nil, nil
		}

		currNode = foundNode
	}

	return currNode, params
}

func (n *node) find(part string) *node {
	var children = n.child
	if part[0] == ':' {
		part = part[1:]
		children = n.wildChild
	}

	for _, child := range children {
		if child.part == part {
			return child
		}
	}
	return nil
}

func (n *node) insert(path string, handler http.HandlerFunc) {
	parts := strings.Split(path[1:], "/")
	currNode := n
	for _, part := range parts {
		child := currNode.find(part)
		if child == nil {
			isWild := false
			path = currNode.path + "/" + part
			if part[0] == ':' {
				isWild = true
				part = part[1:]
				path = currNode.path + "/:" + part
			}
			newNode := &node{
				path:    path,
				part:    part,
				isWild:  isWild,
				handler: nil,
			}
			if isWild {
				currNode.wildChild = append(currNode.wildChild, newNode)
			} else {
				currNode.child = append(currNode.child, newNode)
			}
			child = newNode
		}
		currNode = child
	}
	currNode.handler = handler
}

type Mux struct {
	tries   map[string]*node
	options map[string]string

	notFoundHandler http.Handler
}

func (m *Mux) String() string {
	var b bytes.Buffer
	for k, v := range m.tries {
		b.WriteString(k)
		b.WriteString("\n\r")
		b.WriteString(v.String(0))
	}
	return b.String()
}

func NewMux() *Mux {
	return &Mux{
		tries:   map[string]*node{},
		options: map[string]string{},
	}
}

func (m *Mux) notFound(w http.ResponseWriter, r *http.Request) {
	if nil != m.notFoundHandler {
		m.notFoundHandler.ServeHTTP(w, r)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	w.Header().Add("content-type", "text/plain")
	_, _ = w.Write([]byte("404 - NOT FOUND"))
}

func (m *Mux) NotFoundHandler(handler http.Handler) {
	m.notFoundHandler = handler
}

func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)

		for _, n := range m.tries {
			p, _ := n.search(r.URL.Path[1:])
			if p != nil {
				allowed := m.options[p.path]
				w.Header().Add("Access-Control-Allow-Methods", allowed)
				return
			}
		}
	}

	root, ok := m.tries[r.Method]
	if !ok {
		m.notFound(w, r)
		return
	}

	nh, p := root.search(r.URL.Path[1:])

	if nh != nil && nh.handler != nil {
		ctx := context.WithValue(r.Context(), Value("params"), p)
		r = r.WithContext(ctx)
		nh.handler(w, r)
		return
	}

	m.notFound(w, r)
}

func (m *Mux) Handle(method, path string, handler http.HandlerFunc) {
	method = strings.ToUpper(method)
	root := m.tries[method]
	if root == nil {
		root = &node{path: "", part: "", child: []*node{}}
		m.tries[method] = root
	}
	if _, ok := m.options[path]; ok {
		m.options[path] = m.options[path] + ", " + method
	} else {
		m.options[path] = method
	}

	root.insert(path, handler)
}
