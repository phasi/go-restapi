package restapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDefaultCORSRespectsGlobalSetting(t *testing.T) {
	// Store original setting and restore after tests
	originalSetting := GetCORSAlwaysOn()
	defer SetCORSAlwaysOn(originalSetting)

	t.Run("Router with no CORSConfig - strict mode should not set CORS headers", func(t *testing.T) {
		SetCORSAlwaysOn(false) // Strict mode

		router := &Router{BasePath: "/api"}
		router.HandleFunc("GET", "/test", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			WriteJSON(w, map[string]string{"message": "test"})
		})

		// Request without Origin header
		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should NOT set CORS headers in strict mode without Origin
		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "" {
			t.Errorf("Expected no CORS headers in strict mode without Origin, got Access-Control-Allow-Origin: '%s'", origin)
		}

		if methods := w.Header().Get("Access-Control-Allow-Methods"); methods != "" {
			t.Errorf("Expected no CORS headers in strict mode without Origin, got Access-Control-Allow-Methods: '%s'", methods)
		}
	})

	t.Run("Router with no CORSConfig - always-on mode should set CORS headers", func(t *testing.T) {
		SetCORSAlwaysOn(true) // Always-on mode

		router := &Router{BasePath: "/api"}
		router.HandleFunc("GET", "/test", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			WriteJSON(w, map[string]string{"message": "test"})
		})

		// Request without Origin header
		req := httptest.NewRequest("GET", "/api/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should set CORS headers in always-on mode
		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin: '*' in always-on mode, got '%s'", origin)
		}

		if methods := w.Header().Get("Access-Control-Allow-Methods"); methods == "" {
			t.Error("Expected Access-Control-Allow-Methods to be set in always-on mode")
		}
	})

	t.Run("MultiRouter with no CORSConfig - strict mode should not set CORS headers", func(t *testing.T) {
		SetCORSAlwaysOn(false) // Strict mode

		router := &Router{BasePath: "/users"}
		router.HandleFunc("GET", "/", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			WriteJSON(w, map[string]string{"message": "users"})
		})

		multiRouter, err := NewMultiRouter("/api/v1", []*Router{router})
		if err != nil {
			t.Fatal(err)
		}

		// Request without Origin header
		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		w := httptest.NewRecorder()
		multiRouter.ServeHTTP(w, req)

		// Should NOT set CORS headers in strict mode without Origin
		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "" {
			t.Errorf("Expected no CORS headers in MultiRouter strict mode without Origin, got Access-Control-Allow-Origin: '%s'", origin)
		}
	})

	t.Run("MultiRouter with no CORSConfig - always-on mode should set CORS headers", func(t *testing.T) {
		SetCORSAlwaysOn(true) // Always-on mode

		router := &Router{BasePath: "/users"}
		router.HandleFunc("GET", "/", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			WriteJSON(w, map[string]string{"message": "users"})
		})

		multiRouter, err := NewMultiRouter("/api/v1", []*Router{router})
		if err != nil {
			t.Fatal(err)
		}

		// Request without Origin header
		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		w := httptest.NewRecorder()
		multiRouter.ServeHTTP(w, req)

		// Should set CORS headers in always-on mode
		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("Expected Access-Control-Allow-Origin: '*' in MultiRouter always-on mode, got '%s'", origin)
		}
	})

	t.Run("Both modes should work the same when Origin header is present", func(t *testing.T) {
		for _, alwaysOn := range []bool{true, false} {
			SetCORSAlwaysOn(alwaysOn)

			router := &Router{BasePath: "/api"}
			router.HandleFunc("GET", "/test", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
				WriteJSON(w, map[string]string{"message": "test"})
			})

			// Request WITH Origin header
			req := httptest.NewRequest("GET", "/api/test", nil)
			req.Header.Set("Origin", "https://example.com")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Should always set CORS headers when Origin is present
			if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
				t.Errorf("Mode %v: Expected Access-Control-Allow-Origin: '*' when Origin is present, got '%s'", alwaysOn, origin)
			}
		}
	})
}
