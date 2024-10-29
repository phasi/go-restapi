package restapi

import (
	"fmt"
	"net/http"
	"strings"

	"errors"
)

type Permission uint
type RouteContext struct {
	Params              *RouteParams
	userId              string
	requiredPermissions []Permission
	CustomData          *CustomData
}

func (rc *RouteContext) HasRequiredPermissions(userPermissions []Permission) (hasAllPermissions bool) {
	hasAllPermissions = true
	if (rc.requiredPermissions == nil) || (len(rc.requiredPermissions) == 0) {
		return
	}
	for _, requiredPermission := range rc.requiredPermissions {
		hasPermission := false
		for _, userPermission := range userPermissions {
			if userPermission == requiredPermission {
				hasPermission = true
				break
			}
		}
		if !hasPermission {
			hasAllPermissions = false
			break
		}
	}
	return
}

func (rc *RouteContext) GetRequiredPermissions() ([]Permission, error) {
	if rc.requiredPermissions == nil {
		return nil, errors.New("permissions not set")
	}
	return rc.requiredPermissions, nil
}

func (rc *RouteContext) GetUserId() (string, error) {
	if rc.userId == "" {
		return "", errors.New("userId not set")
	}
	return rc.userId, nil
}

func (rc *RouteContext) SetUserId(userId string) {
	rc.userId = userId
}

type RouteParams map[string]string

func (rp RouteParams) Get(key string) (string, error) {
	value, ok := rp[key]
	if !ok || value == "" {
		return "", fmt.Errorf("parameter %s not found or its value is empty", key)
	}
	return value, nil
}

type CustomData map[string]interface{}

func (cd CustomData) Get(key string) (interface{}, error) {
	value, ok := cd[key]
	if !ok {
		return nil, fmt.Errorf("key %s not found", key)
	}
	return value, nil
}

func (cd CustomData) Set(key string, value interface{}) {
	cd[key] = value
}

type RouteHandlerFunc func(http.ResponseWriter, *http.Request, *RouteContext)

type Route struct {
	Method              string
	RelativePath        string
	RequiredPermissions []Permission
	Handler             RouteHandlerFunc
	Protected           bool
}

type Router struct {
	BasePath                string
	Routes                  []Route
	AuthorizationMiddleware func(context *RouteContext, handler http.Handler) http.Handler
	PermissionMiddleware    func(context *RouteContext, handler http.Handler) http.Handler
	CORSConfig              *CORSConfig
}

func (router *Router) HandleFunc(method, path string, handler RouteHandlerFunc) {
	fixedPath := strings.TrimRight(router.BasePath, "/") + path
	if path == "/" {
		fixedPath = router.BasePath
	}
	route := Route{
		Method:       method,
		RelativePath: fixedPath,
		Handler:      handler,
		Protected:    false,
	}
	router.Routes = append(router.Routes, route)
}

func (router *Router) HandleProtectedFunc(method, path string, requiredPermissions []Permission, handler RouteHandlerFunc) {
	fixedPath := strings.TrimRight(router.BasePath, "/") + path
	if path == "/" {
		fixedPath = router.BasePath
	}
	route := Route{
		Method:              method,
		RelativePath:        fixedPath,
		Handler:             handler,
		RequiredPermissions: requiredPermissions,
		Protected:           true,
	}
	router.Routes = append(router.Routes, route)
}

func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// handle CORS
	if router.CORSConfig == nil {
		origin := req.Header.Get("Origin")
		w.Header().Set("Access-Control-Allow-Origin", origin)
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	} else {
		allowed := router.CORSConfig.HandleCORS(w, req)
		if !allowed {
			w.Write([]byte("CORS not allowed"))
			return
		}
	}
	if req.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	for _, route := range router.Routes {
		if req.Method != route.Method {
			continue
		}
		routeSegments := strings.Split(route.RelativePath, "/")
		pathSegments := strings.Split(req.URL.Path, "/")
		if len(routeSegments) != len(pathSegments) {
			continue
		}
		params := make(RouteParams)
		routeContext := &RouteContext{Params: &params}
		match := true
		for i, routeSegment := range routeSegments {
			if strings.HasPrefix(routeSegment, ":") {
				params[routeSegment[1:]] = pathSegments[i]
			} else if routeSegment != pathSegments[i] {
				match = false
				break
			}
		}
		// pass required permissions to route context
		routeContext.requiredPermissions = route.RequiredPermissions
		// pass custom data to route context
		customData := make(CustomData)
		routeContext.CustomData = &customData

		if match {
			if route.Protected {
				if router.AuthorizationMiddleware == nil {
					http.Error(w, "Router.AuthorizationMiddleware is not set", http.StatusInternalServerError)
					return
				}
				if router.PermissionMiddleware == nil {
					http.Error(w, "Router.PermissionMiddleware is not set", http.StatusInternalServerError)
					return
				}
				router.AuthorizationMiddleware(routeContext, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					router.PermissionMiddleware(routeContext, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						route.Handler(w, r, routeContext)
					})).ServeHTTP(w, r)
				})).ServeHTTP(w, req)
				return
			}
			route.Handler(w, req, routeContext)
			return
		}
	}
	http.NotFound(w, req)
}

type MultiRouter struct {
	BasePath string
	Routers  []*Router
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

	// parse the request path
	pathSegments := strings.Split(req.URL.Path, "/")
	basePath := strings.TrimPrefix(mr.BasePath, "/")
	if basePath != pathSegments[1] {
		http.NotFound(w, req)
		return
	}

	if len(pathSegments) < 3 {
		http.NotFound(w, req)
		return
	}

	// find the router that matches the second path segment
	for _, router := range mr.Routers {
		routerBasePath := strings.TrimSuffix(router.BasePath, "/")
		routerBasePath = strings.TrimPrefix(routerBasePath, "/")
		if routerBasePath == pathSegments[2] {
			// req.URL.Path = "/" + strings.Join(pathSegments[2:], "/")
			router.ServeHTTP(w, req)
			return
		}
	}
	http.NotFound(w, req)
}
