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
	originHeaderMissing := requestOrigin == ""

	// Check if the request origin is in the allowed origins list
	if len(config.AllowedOrigins) > 0 {
		for _, origin := range config.AllowedOrigins {
			if origin == "*" {
				allowedOrigin = "*"
				break
			} else if origin == requestOrigin && !originHeaderMissing {
				allowedOrigin = requestOrigin
				break
			}
		}

		// Handle missing Origin header case
		if originHeaderMissing {
			// If wildcard is allowed, set it for missing origin
			for _, origin := range config.AllowedOrigins {
				if origin == "*" {
					allowedOrigin = "*"
					break
				}
			}
		}
	}

	// Set Access-Control-Allow-Origin based on configuration
	if allowedOrigin != "" {
		w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
	} else if originHeaderMissing && corsAlwaysOn {
		// For missing origin, be permissive with origin but restrictive with credentials
		// Only when corsAlwaysOn is enabled (developer-friendly mode)
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	// If corsAlwaysOn is false and no origin match, don't set CORS headers (spec-compliant)

	// Only set other CORS headers if we're setting Allow-Origin OR if corsAlwaysOn is enabled
	shouldSetCORSHeaders := (w.Header().Get("Access-Control-Allow-Origin") != "") || (corsAlwaysOn && originHeaderMissing)

	if shouldSetCORSHeaders {
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
	}

	// Handle Credentials - only set if we're setting other CORS headers
	if shouldSetCORSHeaders {
		if config.AllowCredentials && !originHeaderMissing {
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
			// Always set credentials to false when:
			// 1. AllowCredentials is false, OR
			// 2. Origin header is missing (security best practice)
			w.Header().Set("Access-Control-Allow-Credentials", "false")

			// Log security notice for missing origin with credentials enabled
			if config.AllowCredentials && originHeaderMissing && corsAlwaysOn {
				// TODO: Add logging here "CORS: Origin header missing - credentials disabled for security\n"
			}
		}
	}

	// Handle Max-Age for preflight requests - only if we're setting CORS headers
	if shouldSetCORSHeaders && config.MaxAge > 0 && r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
	}
}
