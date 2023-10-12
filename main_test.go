package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouter(t *testing.T) {
	router := NewMux()

	router.Handle("GET", "/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "This handles the GET /users")
	})

	router.Handle("POST", "/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "This handles the POST /users")
	})

	router.Handle("PUT", "/users/:id", func(w http.ResponseWriter, r *http.Request) {
		params := r.Context().Value(Value("params")).(map[string]string)
		fmt.Fprintf(w, "This handles the PUT /users/:id -> /users/%s", params["id"])
	})

	router.Handle("DELETE", "/users/:id", func(w http.ResponseWriter, r *http.Request) {
		params := r.Context().Value(Value("params")).(map[string]string)
		fmt.Fprintf(w, "This handles the DELETE /users/:id -> /users/%s", params["id"])
	})

	router.Handle("GET", "/users/:id", func(w http.ResponseWriter, r *http.Request) {
		params := r.Context().Value(Value("params")).(map[string]string)
		fmt.Fprintf(w, "This handles the GET /users/:id -> /users/%s", params["id"])
	})

	tests := []struct {
		method       string
		url          string
		expected     string
		expectedCode int
	}{
		{method: "GET", url: "/user", expected: "404 page not found", expectedCode: http.StatusNotFound},
		{method: "GET", url: "/users", expected: "This handles the GET /users", expectedCode: http.StatusOK},
		{method: "POST", url: "/users", expected: "This handles the POST /users", expectedCode: http.StatusCreated},
		{method: "GET", url: "/users/U1234", expected: "This handles the GET /users/:id -> /users/U1234", expectedCode: http.StatusOK},
		{method: "PUT", url: "/users/U1234", expected: "This handles the PUT /users/:id -> /users/U1234", expectedCode: http.StatusOK},
		{method: "DELETE", url: "/users/U1234", expected: "This handles the DELETE /users/:id -> /users/U1234", expectedCode: http.StatusOK},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s__%s", test.method, test.url), func(t *testing.T) {
			req := httptest.NewRequest(test.method, test.url, nil)
			w := httptest.NewRecorder()

			ctx := context.WithValue(req.Context(), Value("params"), map[string]string{})
			req = req.WithContext(ctx)

			router.ServeHTTP(w, req)

			code := w.Code
			body := w.Body.String()

			if code != test.expectedCode {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, code)
			}

			if body != test.expected {
				t.Errorf("Expected response '%s', got '%s'", test.expected, body)
			}
		})
	}
}
