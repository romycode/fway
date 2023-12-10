package main

//
//import (
//	"fmt"
//	"log"
//	"net/http"
//
//	"github.com/romycode/fway"
//)
//
//func main() {
//	router := fway.NewMux()
//
//	router.Handle("GET", "/users/:id", func(w http.ResponseWriter, r *http.Request) {
//		params := fway.Params(r)
//		fmt.Fprintf(w, "This handles the get /users/:id route -> resolved to: /users/%s", params["id"])
//	})
//
//	router.Handle("PUT", "/users/:id", func(w http.ResponseWriter, r *http.Request) {
//		fmt.Fprint(w, "This handles the put /users/:id route")
//	})
//
//	router.Handle("PUT", "/users/:id/subscriptions/:id", func(w http.ResponseWriter, r *http.Request) {
//		fmt.Fprint(w, "This handles the put /users/:id/subscriptions/:id route")
//	})
//
//	router.Handle("PUT", "/users/:id/subscriptions/:id/products", func(w http.ResponseWriter, r *http.Request) {
//		fmt.Fprint(w, "This handles the put /users/:id/subscriptions/:id/products route")
//	})
//
//	router.Handle("PUT", "/users/:id/subscriptions/:id/finish_date", func(w http.ResponseWriter, r *http.Request) {
//		fmt.Fprint(w, "This handles the put /users/:id/subscriptions/:id/finish_date route")
//	})
//
//	router.CustomNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		w.WriteHeader(http.StatusNotFound)
//	}))
//
//	http.Handle("/", router)
//
//	fmt.Println("Listening on localhost:3333")
//	err := http.ListenAndServe(":3333", nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//}
