package fway

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkMux_ServeHTTP(b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		router := NewMux()
		router.Handle("GET", "/users", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "/users")
		})

		router.Handle("GET", "/us", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "/users")
		})

		router.Handle("GET", "/use", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "/users")
		})

		router.Handle("GET", "/users/1234/12212/sas", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "/users")
		})

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
	}
}

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
			"\n  |-> users [ uri: '/users' handler: true ]" +
			"\n    |-> :id [ uri: '/users/:id' handler: true ]" +
			"\n      |-> subscriptions [ uri: '/users/:id/subscriptions' handler: true ]" +
			"\n"

		actual := n.String(0)

		if expectedNode != actual {
			t.Fatalf("expected:\n%s\ngot:\n%s\n", expectedNode, actual)
		}
	})
}

func TestRouter(t *testing.T) {
	router := NewMux()

	router.Handle("GET", "/users", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "This handles the GET /users")
	})

	router.Handle("GET", "/users/:id", func(w http.ResponseWriter, r *http.Request) {
		params := Params(r)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "This handles the GET /users/:id -> /users/%s", params["id"])
	})

	router.Handle("GET", "/users/:id/subscriptors", func(w http.ResponseWriter, r *http.Request) {
		params := Params(r)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "This handles the GET /users/:id/subscriptors -> /users/%s/subscriptors", params["id"])
	})

	tests := []struct {
		method       string
		url          string
		expected     string
		expectedCode int
	}{
		{method: "GET", url: "/user", expected: "404 - NOT FOUND", expectedCode: http.StatusNotFound},
		{method: "GET", url: "/users", expected: "This handles the GET /users", expectedCode: http.StatusOK},
		{method: "GET", url: "/users/1234", expected: "This handles the GET /users/:id -> /users/1234", expectedCode: http.StatusOK},
		{method: "GET", url: "/users/1234/subscriptors", expected: "This handles the GET /users/:id/subscriptors -> /users/1234/subscriptors", expectedCode: http.StatusOK},
		{method: "OPTIONS", url: "/users/1234/subscriptors", expected: "", expectedCode: http.StatusNoContent},
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
				t.Errorf("Expected status code %d, got %d", test.expectedCode, code)
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

func TestMux_NotFoundHandler(t *testing.T) {
	router := NewMux()

	expectedBody := `{"error":{"code":"404","details":"not found"}}`

	router.NotFoundHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
