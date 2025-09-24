package restapi

// Global CORS configuration
var (
	// corsAlwaysOn determines if CORS headers should be set even when Origin header is missing
	// true: Always set CORS headers (developer-friendly, non-spec-compliant)
	// false: Only set CORS headers when Origin header is present (spec-compliant)
	corsAlwaysOn = false
)

// SetCORSAlwaysOn configures whether CORS headers should always be set, even without Origin header
//
// When true (developer-friendly):
//   - CORS headers are always set, even for non-CORS requests
//   - Access-Control-Allow-Origin defaults to "*" when Origin header is missing
//   - Helps with debugging and testing tools
//
// When false (spec-compliant, default):
//   - CORS headers are only set when Origin header is present
//   - Follows W3C CORS specification strictly
//   - More secure and standards-compliant
func SetCORSAlwaysOn(alwaysOn bool) {
	corsAlwaysOn = alwaysOn
}

// GetCORSAlwaysOn returns the current CORS always-on setting
func GetCORSAlwaysOn() bool {
	return corsAlwaysOn
}
