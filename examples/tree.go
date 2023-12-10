package main

import (
	"fmt"
	"github.com/romycode/fway"
	"net/http"
)

func main() {
	router := fway.NewMux()

	router.Handle("GET", "/users/:id", func(w http.ResponseWriter, r *http.Request) {
		params := fway.Params(r)
		fmt.Fprintf(w, "This handles the get /users/:id route -> resolved to: /users/%s", params["id"])
	})

	router.Handle("PUT", "/users/:id/subscriptions/:id/products", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "This handles the put /users/:id/subscriptions/:id/products route")
	})

	router.Handle("PUT", "/users/:id/subscriptions/:id/finish_date", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "This handles the put /users/:id/subscriptions/:id/finish_date route")
	})

	router.CustomNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))

	router.Handle("PUT", "/users/:id/subscriptions/:id", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "This handles the put /users/:id/subscriptions/:id route")
	})

	fmt.Println(router.String())
}
