package fway

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNode(t *testing.T) {
	t.Run("insert node", func(t *testing.T) {
		var n = &node{
			path:   "",
			part:   "",
			isWild: false,
		}

		n.insert("/users", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("/users"))
		})
		n.insert("/users/:id", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("/users/:id"))
		})
		n.insert("/users/:id/subscriptions", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("/users/:id"))
		})

		expectedNode := "|-> [ uri: '/' ]" +
			"\n  |-> users [ uri: '/users' ]" +
			"\n    |-> :id [ uri: '/users/:id' ]" +
			"\n      |-> subscriptions [ uri: '/users/:id/subscriptions' ]" +
			"\n"

		actual := n.String(0)

		if expectedNode != actual {
			t.Fatalf("expected: %s\ngot: %s", expectedNode, actual)
		}
	})
}

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

	router.Handle("GET", "/users/:id/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		params := r.Context().Value(Value("params")).(map[string]string)
		fmt.Fprintf(w, "This handles the GET /users/:id/subscriptions -> /users/%s/subscriptions", params["id"])
	})

	tests := []struct {
		method       string
		url          string
		expected     string
		expectedCode int
	}{
		{method: "GET", url: "/user", expected: "404 - NOT FOUND", expectedCode: http.StatusNotFound},
		{method: "GET", url: "/users", expected: "This handles the GET /users", expectedCode: http.StatusOK},
		{method: "POST", url: "/users", expected: "This handles the POST /users", expectedCode: http.StatusCreated},
		{method: "GET", url: "/users/U1234", expected: "This handles the GET /users/:id -> /users/U1234", expectedCode: http.StatusOK},
		{method: "GET", url: "/users/U1234/subscriptions", expected: "This handles the GET /users/:id/subscriptions -> /users/U1234/subscriptions", expectedCode: http.StatusOK},
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

func TestRouter_AutomaticOptions(t *testing.T) {
	router := NewMux()

	router.Handle("GET", "/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "This handles the GET /users")
	})

	router.Handle("PUT", "/users/:id", func(w http.ResponseWriter, r *http.Request) {
		params := r.Context().Value(Value("params")).(map[string]string)
		fmt.Fprintf(w, "This handles the PUT /users/:id -> /users/%s", params["id"])
	})

	tests := []struct {
		method          string
		url             string
		optionsExpected string
		expected        string
		expectedCode    int
	}{
		{method: "GET", url: "/users/U1234", optionsExpected: "PUT", expected: "404 - NOT FOUND", expectedCode: http.StatusNotFound},
		{method: "PUT", url: "/users/U1234", optionsExpected: "PUT", expected: "This handles the PUT /users/:id -> /users/U1234", expectedCode: http.StatusOK},
		{method: "GET", url: "/users", optionsExpected: "GET", expected: "This handles the GET /users", expectedCode: http.StatusOK},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s__%s", test.method, test.url), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodOptions, test.url, nil)
			w := httptest.NewRecorder()

			ctx := context.WithValue(req.Context(), Value("params"), map[string]string{})
			req = req.WithContext(ctx)

			router.ServeHTTP(w, req)

			code := w.Code
			allowed := w.Header().Get("Access-Control-Allow-Methods")

			if code != http.StatusNoContent {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, code)
			}

			if test.optionsExpected != allowed {
				t.Errorf("Expected response '%s', got '%s'", test.optionsExpected, allowed)
			}

			req = httptest.NewRequest(test.method, test.url, nil)
			w = httptest.NewRecorder()

			ctx = context.WithValue(req.Context(), Value("params"), map[string]string{})
			req = req.WithContext(ctx)

			router.ServeHTTP(w, req)

			code = w.Code
			allowed = w.Body.String()

			if code != test.expectedCode {
				t.Errorf("Expected status code %d, got %d", test.expectedCode, code)
			}

			if allowed != test.expected {
				t.Errorf("Expected response '%s', got '%s'", test.expected, allowed)
			}
		})
	}
}

func TestMux_CustomNotFoundHandler(t *testing.T) {
	router := NewMux()

	expectedBody := `{"error":{"code":"404","details":"not found"}}`

	router.CustomNotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(expectedBody))
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	ctx := context.WithValue(req.Context(), Value("params"), map[string]string{})
	req = req.WithContext(ctx)

	router.ServeHTTP(w, req)

	code := w.Code
	body := w.Body.String()

	if code != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, code)
	}

	if body != expectedBody {
		t.Errorf("Expected response '%s', got '%s'", expectedBody, body)
	}
}
