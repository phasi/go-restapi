package restapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSLogic(t *testing.T) {
	// Test 1: Default CORS is now secure
	t.Run("Default CORS is secure", func(t *testing.T) {
		router := &Router{}
		router.HandleFunc("GET", "/test", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://any-site.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		origin := w.Header().Get("Access-Control-Allow-Origin")
		credentials := w.Header().Get("Access-Control-Allow-Credentials")
		methods := w.Header().Get("Access-Control-Allow-Methods")

		if origin != "*" {
			t.Errorf("Expected wildcard origin, got: %s", origin)
		}
		if credentials != "false" {
			t.Errorf("Expected credentials to be false by default, got: %s", credentials)
		}
		t.Logf("Good! Default CORS: Origin=%s, Credentials=%s, Methods=%s", origin, credentials, methods)
	})

	// Test 2: CORSConfig validates origins properly
	t.Run("CORSConfig validates requesting origin", func(t *testing.T) {
		router := &Router{
			CORSConfig: &CORSConfig{
				AllowedOrigins: []string{"https://trusted-site.com"},
			},
		}
		router.HandleFunc("GET", "/test", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://malicious-site.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		allowedOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if allowedOrigin != "" {
			t.Errorf("Expected no CORS header for untrusted origin, but got: %s", allowedOrigin)
		}
		t.Logf("Good! Malicious origin %s was rejected", req.Header.Get("Origin"))
	})

	// Test 3: Trusted origin should be allowed
	t.Run("Trusted origin is allowed", func(t *testing.T) {
		router := &Router{
			CORSConfig: &CORSConfig{
				AllowedOrigins: []string{"https://trusted-site.com", "https://another-trusted.com"},
			},
		}
		router.HandleFunc("GET", "/test", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://trusted-site.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		allowedOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if allowedOrigin != "https://trusted-site.com" {
			t.Errorf("Expected trusted origin to be allowed, got: %s", allowedOrigin)
		}
		t.Logf("Good! Trusted origin %s was allowed", allowedOrigin)
	})

	// Test 4: Wildcard origin handling
	t.Run("Wildcard origin allows all", func(t *testing.T) {
		router := &Router{
			CORSConfig: &CORSConfig{
				AllowedOrigins: []string{"*"},
			},
		}
		router.HandleFunc("GET", "/test", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://any-site.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		allowedOrigin := w.Header().Get("Access-Control-Allow-Origin")
		if allowedOrigin != "*" {
			t.Errorf("Expected wildcard origin, got: %s", allowedOrigin)
		}
		t.Logf("Wildcard CORS working: %s", allowedOrigin)
	})

	// Test 5: Credentials with wildcard (security issue)
	t.Run("Credentials with wildcard should be handled safely", func(t *testing.T) {
		router := &Router{
			CORSConfig: &CORSConfig{
				AllowedOrigins:   []string{"*", "https://trusted.com"},
				AllowCredentials: true,
			},
		}
		router.HandleFunc("GET", "/test", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://trusted.com")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		allowedOrigin := w.Header().Get("Access-Control-Allow-Origin")
		credentials := w.Header().Get("Access-Control-Allow-Credentials")

		if allowedOrigin == "*" && credentials == "true" {
			t.Error("SECURITY ISSUE: Cannot use wildcard origin with credentials!")
		}
		t.Logf("Origin: %s, Credentials: %s", allowedOrigin, credentials)
	})

	// Test 6: OPTIONS request handling
	t.Run("OPTIONS preflight request", func(t *testing.T) {
		router := &Router{
			CORSConfig: &CORSConfig{
				AllowedOrigins: []string{"https://example.com"},
				MaxAge:         3600,
			},
		}
		router.HandleFunc("POST", "/api/data", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			w.WriteHeader(http.StatusOK)
		})

		req := httptest.NewRequest("OPTIONS", "/api/data", nil)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 for OPTIONS, got: %d", w.Code)
		}

		maxAge := w.Header().Get("Access-Control-Max-Age")
		if maxAge != "3600" {
			t.Errorf("Expected Max-Age 3600, got: %s", maxAge)
		}
		t.Logf("OPTIONS request handled with status: %d, Max-Age: %s", w.Code, maxAge)
	})
}
