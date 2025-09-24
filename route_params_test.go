package restapi

import (
	"testing"
)

func TestRouteParamsIssues(t *testing.T) {
	// Test potential issue: RouteParams doesn't have a Set method
	t.Run("RouteParams missing Set method", func(t *testing.T) {
		params := make(RouteParams)
		// Manual setting (set manually since there's no Set method. Set method is only used by internal router logic)
		params["user_id"] = "123"
		userId, err := params.Get("user_id")
		if err != nil || userId != "123" {
			t.Errorf("Expected user_id '123', got: %s, error: %v", userId, err)
		}

	})

}
