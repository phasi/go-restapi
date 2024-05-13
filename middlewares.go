package restapi

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// statusWriter is a wrapper around the ResponseWriter that stores the status code
type statusWriter struct {
	http.ResponseWriter
	status int
}

// WriteHeader is a wrapper around the ResponseWriter's WriteHeader method that stores the status code
func (sw *statusWriter) WriteHeader(statusCode int) {
	sw.status = statusCode
	sw.ResponseWriter.WriteHeader(statusCode)
}

// LoggingRouter is a middleware that logs the request method, URL path and response status code
func LoggingRouter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sw := statusWriter{ResponseWriter: w}
		next.ServeHTTP(&sw, r)
		log.Println(r.Method, r.URL.Path, sw.status)
	})
}

// TracingRouter is a middleware that adds a trace ID to the request context and response headers
func TracingRouter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.New().String()
		ctx := context.WithValue(r.Context(), "traceID", traceID)
		w.Header().Set("X-Trace-ID", traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CORSConfig is a configuration struct for the CORS middleware
type CORSConfig struct {
	// AllowedOrigins is a list of origins allowed to make requests
	AllowedOrigins []string
	// AllowedMethods is a list of HTTP methods allowed in the request
	AllowedMethods []string
	// AllowedHeaders is a list of headers allowed in the request
	AllowedHeaders []string
	// AllowCredentials is a boolean that determines if credentials are allowed in the request
	AllowCredentials bool
	// if User-Agent contains any of the strings in BlockUserAgents, the request will be blocked
	BlockUserAgents []string
}

func (config *CORSConfig) applyCORS(w http.ResponseWriter, r *http.Request) (err error) {
	origin := r.Header.Get("Origin")
	// check if the origin is in allowed origins
	if len(config.AllowedOrigins) > 0 {
		var isAllowedOrigin bool = false
		for _, allowedOrigin := range config.AllowedOrigins {
			if allowedOrigin == "*" || allowedOrigin == origin {
				isAllowedOrigin = true
				break
			}
		}
		if !isAllowedOrigin {
			err = errors.New("Origin not allowed")
			return
		}
	}

	method := r.Method
	if !strings.Contains(strings.Join(config.AllowedMethods, ","), method) {
		err = errors.New("Method not allowed")
		return
	}
	if len(config.AllowedHeaders) > 0 {
		for headerName, headers := range r.Header {
			// Convert header name to lower case for case insensitive comparison
			lowerHeaderName := strings.ToLower(headerName)

			// Block requests with blocked User-Agent
			if lowerHeaderName == "user-agent" {
				for _, blockedUserAgent := range config.BlockUserAgents {
					if strings.Contains(strings.Join(headers, " "), blockedUserAgent) {
						err = errors.New("User-Agent blocked")
						return
					}
				}
			}
			// Allow some headers to be passed through
			if lowerHeaderName == "user-agent" || lowerHeaderName == "accept" || lowerHeaderName == "host" {
				continue
			}

			// Check if the header name is in the list of allowed headers
			if !strings.Contains(strings.ToLower(strings.Join(config.AllowedHeaders, ",")), lowerHeaderName) {
				err = errors.New("Header not allowed")
				return
			}
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ","))
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ","))
	var allowCredentials string
	if config.AllowCredentials == true {
		allowCredentials = "true"
	} else {
		allowCredentials = "false"
	}
	w.Header().Set("Access-Control-Allow-Credentials", allowCredentials)
	return
}

// MiddlewareFunc is a middleware that should be used to wrap individual handler functions (RouteHandlerFunc)
func (config *CORSConfig) MiddlewareFunc(next RouteHandlerFunc) RouteHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, context RouteContext) {
		err := config.applyCORS(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		next(w, r, context)
	}
}

// CORSRouter is a middleware that should be used to wrap the main router
func (config *CORSConfig) CORSRouter(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := config.applyCORS(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}
