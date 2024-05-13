package restapi

import (
	"fmt"
	"net/http"
	"strings"
)

type RouteContext struct {
	Params *RouteParams
	userId string
}

func (rc *RouteContext) GetUserId() string {
	return rc.userId
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

func (rp RouteParams) Set(key, value string) {
	rp[key] = value
}

type RouteHandlerFunc func(http.ResponseWriter, *http.Request, RouteContext)

type Route struct {
	Method       string
	RelativePath string
	Handler      RouteHandlerFunc
	Protected    bool
}

type Router struct {
	BasePath                string
	Routes                  []Route
	AuthorizationMiddleware func(context *RouteContext, handler http.Handler) http.Handler
}

func (r *Router) HandleFunc(method, path string, handler RouteHandlerFunc) {
	route := Route{
		Method:       method,
		RelativePath: strings.TrimRight(r.BasePath, "/") + path,
		Handler:      handler,
		Protected:    false,
	}
	r.Routes = append(r.Routes, route)
}

func (r *Router) HandleFuncProtected(method, path string, handler RouteHandlerFunc) {
	route := Route{
		Method:       method,
		RelativePath: strings.TrimRight(r.BasePath, "/") + path,
		Handler:      handler,
		Protected:    true,
	}
	r.Routes = append(r.Routes, route)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	for _, route := range r.Routes {
		if req.Method != route.Method {
			continue
		}
		routeSegments := strings.Split(route.RelativePath, "/")
		pathSegments := strings.Split(req.URL.Path, "/")
		if len(routeSegments) != len(pathSegments) {
			continue
		}
		params := make(RouteParams)
		context := &RouteContext{Params: &params}
		match := true
		for i, routeSegment := range routeSegments {
			if strings.HasPrefix(routeSegment, ":") {
				params[routeSegment[1:]] = pathSegments[i]
			} else if routeSegment != pathSegments[i] {
				match = false
				break
			}
		}
		if match {
			if route.Protected {
				if r.AuthorizationMiddleware == nil {
					http.Error(w, "Router.AuthorizationMiddleware is not set", http.StatusInternalServerError)
					return
				}
				r.AuthorizationMiddleware(context, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					route.Handler(w, r, *context)
				})).ServeHTTP(w, req)
				return
			}
			route.Handler(w, req, *context)
			return
		}
	}
	http.NotFound(w, req)
}
