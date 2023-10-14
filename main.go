package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
)

type Value string

type node struct {
	child   []*node
	path    string
	part    string
	isWild  bool
	handler http.HandlerFunc
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

			if child.isWild {
				params[child.part] = part
				foundNode = child
				continue
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
			currNode.child = append(currNode.child, newNode)
			sort.Slice(currNode.child, func(i, j int) bool {
				if currNode.child[i].isWild && !currNode.child[j].isWild {
					return false
				}
				return true
			})
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

func main() {
	router := NewMux()

	router.Handle("GET", "/users/:id", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "This handles the get /users/:id route")
	})

	router.Handle("PUT", "/users/:id", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "This handles the put /users/:id route")
	})

	http.Handle("/", router)

	log.Fatal(http.ListenAndServe(":3333", nil))
}
