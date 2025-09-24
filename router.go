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
	// Handle CORS only if not already handled (e.g., by MultiRouter)
	corsAlreadyHandled := w.Header().Get("Access-Control-Allow-Origin") != ""

	if !corsAlreadyHandled {
		// handle CORS
		if router.CORSConfig == nil {
			// Default: restrictive CORS policy for security
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "false")
		} else {
			router.CORSConfig.HandleCORS(w, req)
		}

		if req.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
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
