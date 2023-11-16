package fway

import (
	"context"
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

func (n *node) search(path string) (*node, map[string]string) {
	params := map[string]string{}
	parts := strings.Split(path, "/")

	var currNode = n
	for _, part := range parts {
		var foundNode *node = nil

		for _, child := range currNode.child {
			if child.part == part {
				foundNode = child
				continue
			}
		}

		if foundNode == nil {
			for _, child := range currNode.wildChild {
				if child.isWild {
					foundNode = child
					params[child.part] = part
					continue
				}
			}
		}

		if foundNode == nil {
			return nil, nil
		}

		currNode = foundNode
	}

	if currNode == n {
		return nil, nil
	}

	return currNode, params
}

func (n *node) find(part string) *node {
	for _, child := range n.child {
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
				path:   path,
				part:   part,
				isWild: isWild,
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

func NewMux() *Mux {
	return &Mux{
		tries:   map[string]*node{},
		options: map[string]string{},
	}
}

func (t *Mux) notFound(w http.ResponseWriter, r *http.Request) {
	if nil != t.notFoundHandler {
		t.notFoundHandler.ServeHTTP(w, r)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 - NOT FOUND"))
}

func (t *Mux) CustomNotFoundHandler(handler http.Handler) {
	t.notFoundHandler = handler
}

func (t *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)

		for _, n := range t.tries {
			p, _ := n.search(r.URL.Path[1:])
			if p != nil {
				allowed := t.options[p.path]
				w.Header().Add("Access-Control-Allow-Methods", allowed)
				return
			}
		}
	}

	root, ok := t.tries[r.Method]
	if !ok {
		t.notFound(w, r)
		return
	}

	nh, p := root.search(r.URL.Path[1:])

	if nh != nil {
		ctx := context.WithValue(r.Context(), Value("params"), p)
		r = r.WithContext(ctx)
		nh.handler(w, r)
		return
	}

	t.notFound(w, r)
}

func (t *Mux) Handle(method, path string, handler http.HandlerFunc) {
	method = strings.ToUpper(method)
	root := t.tries[method]
	if root == nil {
		root = &node{child: []*node{}}
		t.tries[method] = root
	}
	if _, ok := t.options[path]; ok {
		t.options[path] = t.options[path] + ", " + method
	} else {
		t.options[path] = method
	}

	root.insert(path, handler)
}
