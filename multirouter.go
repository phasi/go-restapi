package restapi

import (
	"errors"
	"net/http"
	"strings"
)

type MultiRouter struct {
	BasePath   string
	Routers    []*Router
	CORSConfig *CORSConfig
}

// NewMultiRouter is a constructor function for MultiRouter
func NewMultiRouter(basePath string, routers []*Router) (*MultiRouter, error) {
	if basePath == "" || basePath == "/" {
		return nil, errors.New("basePath cannot be empty or '/' for MultiRouter. If you want to use '/' as basePath, use a single Router instead")
	}

	// reconfigure router routes
	for _, router := range routers {
		for i, route := range router.Routes {
			router.Routes[i].RelativePath = basePath + route.RelativePath
		}
	}

	return &MultiRouter{
		BasePath: basePath,
		Routers:  routers,
	}, nil
}

// NewMultiRouterWithCORS creates a MultiRouter with CORS configuration
// This will override all individual router CORS settings
func NewMultiRouterWithCORS(basePath string, routers []*Router, corsConfig *CORSConfig) (*MultiRouter, error) {
	mr, err := NewMultiRouter(basePath, routers)
	if err != nil {
		return nil, err
	}

	// Clear CORS config from individual routers to avoid duplication
	for _, router := range routers {
		router.CORSConfig = nil
	}

	mr.CORSConfig = corsConfig
	return mr, nil
}

func (mr *MultiRouter) ListRoutes() []string {
	var routes []string
	for _, router := range mr.Routers {
		for _, route := range router.Routes {
			routes = append(routes, route.Method+" "+route.RelativePath)
		}
	}
	return routes
}

func (mr *MultiRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Check if the request path starts with the base path
	basePath := strings.TrimSuffix(mr.BasePath, "/")
	if !strings.HasPrefix(req.URL.Path, basePath) {
		http.NotFound(w, req)
		return
	}

	// Find which router should handle this request
	var matchingRouter *Router
	var routeFound bool

	for _, router := range mr.Routers {
		for _, route := range router.Routes {
			routeSegments := strings.Split(route.RelativePath, "/")
			pathSegments := strings.Split(req.URL.Path, "/")
			if len(routeSegments) == len(pathSegments) {
				match := true
				for i, routeSegment := range routeSegments {
					if strings.HasPrefix(routeSegment, ":") {
						// Parameter match - always matches
						continue
					} else if routeSegment != pathSegments[i] {
						match = false
						break
					}
				}
				if match {
					matchingRouter = router
					// For OPTIONS requests, check if this path would match any method
					if req.Method == "OPTIONS" {
						routeFound = true
						break
					}
					// For non-OPTIONS requests, also check method
					if req.Method == route.Method {
						routeFound = true
						break
					}
				}
			}
		}
		if routeFound {
			break
		}
	}

	if !routeFound {
		http.NotFound(w, req)
		return
	}

	// Handle CORS - either at MultiRouter level or per-router level
	if mr.CORSConfig != nil {
		// MultiRouter-level CORS overrides individual router CORS
		mr.CORSConfig.HandleCORS(w, req)
		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	} else if matchingRouter != nil {
		// Per-router CORS handling
		if matchingRouter.CORSConfig == nil {
			// Default CORS for this router
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "false")
		} else {
			matchingRouter.CORSConfig.HandleCORS(w, req)
		}

		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	// Forward the request to the matching router
	if matchingRouter != nil {
		matchingRouter.ServeHTTP(w, req)
		return
	}

	http.NotFound(w, req)
}
