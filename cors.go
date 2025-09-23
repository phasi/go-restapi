package restapi

import (
	"fmt"
	"net/http"
	"strings"
)

// CORSConfig is a configuration struct for the CORS middleware
type CORSConfig struct {
	// AllowedOrigins is a list of origins allowed to make requests
	// Use ["*"] to allow all origins (not recommended for production with credentials)
	AllowedOrigins []string
	// AllowedMethods is a list of HTTP methods allowed in the request
	AllowedMethods []string
	// AllowedHeaders is a list of headers allowed in the request
	AllowedHeaders []string
	// AllowCredentials is a boolean that determines if credentials are allowed in the request
	AllowCredentials bool
	// MaxAge is the maximum age for preflight requests (in seconds)
	MaxAge int
}

func (config *CORSConfig) HandleCORS(w http.ResponseWriter, r *http.Request) {
	// Handle Origin
	requestOrigin := r.Header.Get("Origin")
	allowedOrigin := ""

	// Check if the request origin is in the allowed origins list
	if len(config.AllowedOrigins) > 0 {
		for _, origin := range config.AllowedOrigins {
			if origin == "*" {
				allowedOrigin = "*"
				break
			} else if origin == requestOrigin {
				allowedOrigin = requestOrigin
				break
			}
		}
	}

	// Set the appropriate origin header
	if allowedOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
	}

	// Handle Methods
	if len(config.AllowedMethods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ","))
	} else {
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	}

	// Handle Headers
	if len(config.AllowedHeaders) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ","))
	} else {
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	}

	// Handle Credentials
	if config.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		// Note: When credentials are true, origin cannot be "*"
		if allowedOrigin == "*" {
			// Override to be safe - use the actual request origin if it's in allowed list
			if requestOrigin != "" {
				for _, origin := range config.AllowedOrigins {
					if origin == requestOrigin {
						w.Header().Set("Access-Control-Allow-Origin", requestOrigin)
						break
					}
				}
			}
		}
	} else {
		w.Header().Set("Access-Control-Allow-Credentials", "false")
	}

	// Handle Max-Age for preflight requests
	if config.MaxAge > 0 && r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
	}
}
