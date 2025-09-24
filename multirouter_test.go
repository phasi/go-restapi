package restapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMultiRouterCORS(t *testing.T) {
	// Store original setting and restore after tests
	originalSetting := GetCORSAlwaysOn()
	defer SetCORSAlwaysOn(originalSetting)

	// Helper function to create fresh routers for each test
	createRouters := func() (*Router, *Router) {
		userRouter := &Router{BasePath: "/users"}
		userRouter.HandleFunc("GET", "/", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			WriteJSON(w, map[string]string{"message": "users"})
		})

		orderRouter := &Router{BasePath: "/orders"}
		orderRouter.HandleFunc("GET", "/", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			WriteJSON(w, map[string]string{"message": "orders"})
		})

		return userRouter, orderRouter
	}

	t.Run("MultiRouter with default CORS", func(t *testing.T) {
		SetCORSAlwaysOn(true) // Enable always-on mode for this test

		userRouter, orderRouter := createRouters()
		multiRouter, err := NewMultiRouter("/api/v1", []*Router{userRouter, orderRouter})
		if err != nil {
			t.Fatal(err)
		}

		// Test OPTIONS request (use correct path without trailing slash)
		req := httptest.NewRequest("OPTIONS", "/api/v1/users", nil)
		w := httptest.NewRecorder()
		multiRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
		}

		// Check CORS headers - should be present in always-on mode even without Origin
		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("Expected origin '*', got '%s'", origin)
		}

		if credentials := w.Header().Get("Access-Control-Allow-Credentials"); credentials != "false" {
			t.Errorf("Expected credentials 'false', got '%s'", credentials)
		}
	})

	t.Run("MultiRouter with custom CORS", func(t *testing.T) {
		SetCORSAlwaysOn(false) // Can use strict mode since we're providing Origin header

		userRouter, orderRouter := createRouters()
		corsConfig := &CORSConfig{
			AllowedOrigins:   []string{"https://myapp.com"},
			AllowedMethods:   []string{"GET", "POST"},
			AllowedHeaders:   []string{"Content-Type"},
			AllowCredentials: true,
			MaxAge:           7200,
		}

		multiRouter, err := NewMultiRouterWithCORS("/api/v1", []*Router{userRouter, orderRouter}, corsConfig)
		if err != nil {
			t.Fatal(err)
		}

		// Test OPTIONS request with valid origin (use correct path without trailing slash)
		req := httptest.NewRequest("OPTIONS", "/api/v1/users", nil)
		req.Header.Set("Origin", "https://myapp.com")
		w := httptest.NewRecorder()
		multiRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for OPTIONS, got %d", w.Code)
		}

		// Check CORS headers
		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "https://myapp.com" {
			t.Errorf("Expected origin 'https://myapp.com', got '%s'", origin)
		}

		if credentials := w.Header().Get("Access-Control-Allow-Credentials"); credentials != "true" {
			t.Errorf("Expected credentials 'true', got '%s'", credentials)
		}

		if maxAge := w.Header().Get("Access-Control-Max-Age"); maxAge != "7200" {
			t.Errorf("Expected max-age '7200', got '%s'", maxAge)
		}
	})

	t.Run("Actual route requests still work", func(t *testing.T) {
		SetCORSAlwaysOn(true) // Set to always-on mode for this test

		userRouter, orderRouter := createRouters()
		multiRouter, err := NewMultiRouter("/api/v1", []*Router{userRouter, orderRouter})
		if err != nil {
			t.Fatal(err)
		}

		// Debug: print available routes
		routes := multiRouter.ListRoutes()
		t.Logf("Available routes: %v", routes)

		// Test GET request (note: no trailing slash to match route "/users" + "/" = "/users")
		req := httptest.NewRequest("GET", "/api/v1/users", nil)
		w := httptest.NewRecorder()
		multiRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for GET, got %d. Available routes: %v", w.Code, routes)
		}

		// Should still have CORS headers
		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("Expected origin '*', got '%s'", origin)
		}
	})

	t.Run("Individual routers don't duplicate CORS when using NewMultiRouterWithCORS", func(t *testing.T) {
		// Create a router with its own CORS config
		routerWithCORS := &Router{
			BasePath: "/test",
			CORSConfig: &CORSConfig{
				AllowedOrigins: []string{"https://should-be-ignored.com"},
			},
		}
		routerWithCORS.HandleFunc("GET", "/", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			WriteJSON(w, map[string]string{"message": "test"})
		})

		corsConfig := &CORSConfig{
			AllowedOrigins: []string{"https://multirouter-cors.com"},
		}

		multiRouter, err := NewMultiRouterWithCORS("/api/v1", []*Router{routerWithCORS}, corsConfig)
		if err != nil {
			t.Fatal(err)
		}

		// The individual router's CORS should be cleared
		if routerWithCORS.CORSConfig != nil {
			t.Error("Individual router's CORS config should be cleared")
		}

		// Test that MultiRouter CORS takes precedence
		req := httptest.NewRequest("OPTIONS", "/api/v1/test", nil)
		req.Header.Set("Origin", "https://multirouter-cors.com")
		w := httptest.NewRecorder()
		multiRouter.ServeHTTP(w, req)

		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "https://multirouter-cors.com" {
			t.Errorf("Expected MultiRouter CORS to take precedence, got origin '%s'", origin)
		}
	})

	t.Run("Per-router CORS with different settings", func(t *testing.T) {
		// Create public API router with permissive CORS
		publicRouter := &Router{
			BasePath: "/public",
			CORSConfig: &CORSConfig{
				AllowedOrigins:   []string{"*"},
				AllowedMethods:   []string{"GET", "POST"},
				AllowCredentials: false,
			},
		}
		publicRouter.HandleFunc("GET", "/data", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			WriteJSON(w, map[string]string{"type": "public"})
		})

		// Create private API router with restrictive CORS
		privateRouter := &Router{
			BasePath: "/private",
			CORSConfig: &CORSConfig{
				AllowedOrigins:   []string{"https://internal-app.com"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
				AllowCredentials: true,
				MaxAge:           7200,
			},
		}
		privateRouter.HandleFunc("GET", "/data", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			WriteJSON(w, map[string]string{"type": "private"})
		})

		// Create admin API router with very restrictive CORS
		adminRouter := &Router{
			BasePath: "/admin",
			CORSConfig: &CORSConfig{
				AllowedOrigins:   []string{"https://admin.internal.com"},
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
				AllowCredentials: true,
				MaxAge:           3600,
			},
		}
		adminRouter.HandleFunc("GET", "/users", func(w http.ResponseWriter, r *http.Request, ctx *RouteContext) {
			WriteJSON(w, map[string]string{"type": "admin"})
		})

		// Use NewMultiRouter to preserve individual CORS settings (default behavior)
		multiRouter, err := NewMultiRouter("/api/v1", []*Router{publicRouter, privateRouter, adminRouter})
		if err != nil {
			t.Fatal(err)
		}

		// Test public API CORS (should allow any origin)
		req := httptest.NewRequest("OPTIONS", "/api/v1/public/data", nil)
		req.Header.Set("Origin", "https://random-website.com")
		w := httptest.NewRecorder()
		multiRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for public OPTIONS, got %d", w.Code)
		}
		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "*" {
			t.Errorf("Expected public API to allow any origin, got '%s'", origin)
		}
		if credentials := w.Header().Get("Access-Control-Allow-Credentials"); credentials != "false" {
			t.Errorf("Expected public API credentials 'false', got '%s'", credentials)
		}

		// Test private API CORS (should only allow internal-app.com)
		req = httptest.NewRequest("OPTIONS", "/api/v1/private/data", nil)
		req.Header.Set("Origin", "https://internal-app.com")
		w = httptest.NewRecorder()
		multiRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for private OPTIONS, got %d", w.Code)
		}
		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "https://internal-app.com" {
			t.Errorf("Expected private API to allow internal-app.com, got '%s'", origin)
		}
		if credentials := w.Header().Get("Access-Control-Allow-Credentials"); credentials != "true" {
			t.Errorf("Expected private API credentials 'true', got '%s'", credentials)
		}
		if maxAge := w.Header().Get("Access-Control-Max-Age"); maxAge != "7200" {
			t.Errorf("Expected private API max-age '7200', got '%s'", maxAge)
		}

		// Test private API with wrong origin (should be rejected)
		req = httptest.NewRequest("OPTIONS", "/api/v1/private/data", nil)
		req.Header.Set("Origin", "https://malicious-site.com")
		w = httptest.NewRecorder()
		multiRouter.ServeHTTP(w, req)

		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin == "https://malicious-site.com" {
			t.Error("Private API should reject malicious origins")
		}

		// Test admin API CORS (should only allow admin.internal.com)
		req = httptest.NewRequest("OPTIONS", "/api/v1/admin/users", nil)
		req.Header.Set("Origin", "https://admin.internal.com")
		w = httptest.NewRecorder()
		multiRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for admin OPTIONS, got %d", w.Code)
		}
		if origin := w.Header().Get("Access-Control-Allow-Origin"); origin != "https://admin.internal.com" {
			t.Errorf("Expected admin API to allow admin.internal.com, got '%s'", origin)
		}
		if maxAge := w.Header().Get("Access-Control-Max-Age"); maxAge != "3600" {
			t.Errorf("Expected admin API max-age '3600', got '%s'", maxAge)
		}
	})
}
