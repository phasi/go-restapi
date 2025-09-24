# Go REST API Framework

A lightweight, feature-rich REST API framework for Go with built-in routing, middleware, authentication, permissions, CORS, and JSON utilities.

## Features

- 🚀 **Fast HTTP Router** with path parameters
- 🔐 **Built-in Authentication & Authorization** middleware
- 🌐 **CORS Support** with flexible configuration
- 📝 **JSON Utilities** for request/response handling
- 📊 **Logging & Tracing** middleware
- 🎯 **Permission-based Access Control**
- 🔗 **Multi-Router Support** for complex applications
- ⚡ **Zero Dependencies** (except `github.com/google/uuid`)

## Installation

```bash
go get github.com/your-username/go-restapi
```

## Quick Start

```go
package main

import (
    "fmt"
    "net/http"
    api "github.com/your-username/go-restapi"
)

func main() {
    // Create a new router
    router := &api.Router{BasePath: "/api/v1"}

    // Add public routes
    router.HandleFunc("GET", "/users/:id", func(w http.ResponseWriter, r *http.Request, ctx *api.RouteContext) {
        id, _ := ctx.Params.Get("id")
        api.WriteJSON(w, map[string]string{"user_id": id})
    })

    // Start server
    http.ListenAndServe(":8080", router)
}
```

## Core Concepts

### Router

The `Router` is the main component that handles HTTP requests and routes them to appropriate handlers.

```go
router := &api.Router{
    BasePath: "/api/v1", // Optional base path for all routes
}

// Add routes
router.HandleFunc("GET", "/users", getUsersHandler)
router.HandleFunc("POST", "/users", createUserHandler)
router.HandleFunc("PUT", "/users/:id", updateUserHandler)
router.HandleFunc("DELETE", "/users/:id", deleteUserHandler)
```

### Route Parameters

Extract dynamic segments from URLs using the `:parameter` syntax:

```go
router.HandleFunc("GET", "/users/:id/posts/:postId", func(w http.ResponseWriter, r *http.Request, ctx *api.RouteContext) {
    userId, err := ctx.Params.Get("id")
    if err != nil {
        http.Error(w, "Invalid user ID", http.StatusBadRequest)
        return
    }

    postId, err := ctx.Params.Get("postId")
    if err != nil {
        http.Error(w, "Invalid post ID", http.StatusBadRequest)
        return
    }

    // Use userId and postId...
})
```

## Authentication & Authorization

### Define Permissions

First, define your application's permissions:

```go
// Define your permission constants
const (
    PermissionViewUsers   api.Permission = 1
    PermissionEditUsers   api.Permission = 2
    PermissionDeleteUsers api.Permission = 3
    PermissionManageRoles api.Permission = 10
    PermissionSuperAdmin  api.Permission = 99
)
```

### Set Up Middleware

Configure authentication and permission middleware:

```go
router := &api.Router{BasePath: "/api/v1"}

// Authentication middleware - validates tokens and sets user ID
router.AuthorizationMiddleware = func(context *api.RouteContext, handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")

        // Validate token (your implementation)
        if user := validateToken(token); user != nil {
            context.SetUserId(user.ID)
            // Optionally add user data to params
            context.Params.Set("user_role", user.Role)
            handler.ServeHTTP(w, r)
        } else {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
        }
    })
}

// Permission middleware - checks user permissions
router.PermissionMiddleware = func(context *api.RouteContext, handler http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userId, _ := context.GetUserId()

        // Get user permissions from database (your implementation)
        userPermissions := getUserPermissions(userId)

        if context.HasRequiredPermissions(userPermissions) {
            handler.ServeHTTP(w, r)
        } else {
            http.Error(w, "Forbidden", http.StatusForbidden)
        }
    })
}
```

### Protected Routes

Create routes that require authentication and specific permissions:

```go
// Public route - no authentication required
router.HandleFunc("GET", "/health", healthCheckHandler)

// Protected route - requires authentication and permissions
router.HandleProtectedFunc("GET", "/admin/users",
    []api.Permission{PermissionViewUsers},
    func(w http.ResponseWriter, r *http.Request, ctx *api.RouteContext) {
        userId, _ := ctx.GetUserId()
        // Only authenticated users with PermissionViewUsers can access this
        api.WriteJSON(w, map[string]string{"admin": userId})
    })

// Multiple permissions required
router.HandleProtectedFunc("DELETE", "/admin/users/:id",
    []api.Permission{PermissionDeleteUsers, PermissionSuperAdmin},
    deleteUserHandler)
```

## CORS Configuration

### Default CORS (Secure)

By default, the router applies secure CORS headers:

```go
router := &api.Router{} // Default CORS is applied
// Access-Control-Allow-Origin: *
// Access-Control-Allow-Credentials: false
// Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
// Access-Control-Allow-Headers: Content-Type, Authorization
```

### Custom CORS

Configure CORS for your specific needs:

```go
router := &api.Router{
    CORSConfig: &api.CORSConfig{
        AllowedOrigins:   []string{"https://myapp.com", "https://admin.myapp.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Custom-Header"},
        AllowCredentials: true,
        MaxAge:           3600, // Cache preflight for 1 hour
    },
}
```

### CORS Examples

```go
// Allow all origins (development only)
CORSConfig: &api.CORSConfig{
    AllowedOrigins: []string{"*"},
    AllowCredentials: false, // Must be false with "*"
}

// Production setup
CORSConfig: &api.CORSConfig{
    AllowedOrigins: []string{
        "https://myapp.com",
        "https://staging.myapp.com",
    },
    AllowedMethods: []string{"GET", "POST", "PUT", "DELETE"},
    AllowCredentials: true,
    MaxAge: 86400, // 24 hours
}
```

## JSON Utilities

### Writing JSON Responses

```go
// With default response template (includes timestamp)
func getUserHandler(w http.ResponseWriter, r *http.Request, ctx *api.RouteContext) {
    user := User{ID: 1, Name: "John Doe"}
    api.WriteJSON(w, user)
    // Output: {"timestamp": 1640995200, "data": {"id": 1, "name": "John Doe"}}
}

// Without template (raw JSON)
func getRawUserHandler(w http.ResponseWriter, r *http.Request, ctx *api.RouteContext) {
    user := User{ID: 1, Name: "John Doe"}
    api.WriteJSONWithoutTemplate(w, user)
    // Output: {"id": 1, "name": "John Doe"}
}

// Custom response template
func init() {
    api.SetJSONResponseFormatter(func(data interface{}) interface{} {
        return map[string]interface{}{
            "success": true,
            "result":  data,
            "version": "v1",
        }
    })
}
```

### Reading JSON Requests

```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func createUserHandler(w http.ResponseWriter, r *http.Request, ctx *api.RouteContext) {
    var req CreateUserRequest
    if err := api.ReadJSON(r, &req); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    // Process the request...
    user := createUser(req.Name, req.Email)
    api.WriteJSON(w, user)
}
```

## Middleware

### Logging Middleware

Log all HTTP requests with customizable output:

```go
// Custom log function
logFunc := func(entry api.HttpLogEntry) {
    fmt.Printf("[%s] %s %d - %s\\n",
        entry.Method, entry.Path, entry.Status, entry.TraceID)
}

// Apply logging middleware
loggedRouter := api.LoggingRouter(router, logFunc)

// Redact sensitive headers
api.SetRedactedHeaderNames([]string{"Authorization", "X-API-Key"})
```

### Tracing Middleware

Add trace IDs to requests and responses:

```go
// Apply tracing middleware
tracedRouter := api.TracingRouter(router)
// Adds X-Trace-ID header to responses
// Adds trace ID to request context
```

### Chain Middlewares

```go
// Chain multiple middlewares
router := &api.Router{BasePath: "/api/v1"}
loggedRouter := api.LoggingRouter(router, logFunc)
tracedRouter := api.TracingRouter(loggedRouter)

http.ListenAndServe(":8080", tracedRouter)
```

## Multi-Router Support

For complex applications with multiple API versions or modules:

```go
// Create separate routers for different concerns
userRouter := &api.Router{BasePath: "/users"}
userRouter.HandleFunc("GET", "/", listUsersHandler)
userRouter.HandleFunc("POST", "/", createUserHandler)
userRouter.HandleFunc("GET", "/:id", getUserHandler)

orderRouter := &api.Router{BasePath: "/orders"}
orderRouter.HandleFunc("GET", "/", listOrdersHandler)
orderRouter.HandleFunc("POST", "/", createOrderHandler)

// Combine them under a common base path
multiRouter, err := api.NewMultiRouter("/api/v1", []*Router{userRouter, orderRouter})
if err != nil {
    log.Fatal(err)
}

// Routes become:
// GET  /api/v1/users/
// POST /api/v1/users/
// GET  /api/v1/users/:id
// GET  /api/v1/orders/
// POST /api/v1/orders/
```

### Custom Data in Context

Store and retrieve custom data during request processing:

```go
router.HandleFunc("GET", "/example", func(w http.ResponseWriter, r *http.Request, ctx *api.RouteContext) {
    // Set custom data
    ctx.CustomData.Set("start_time", time.Now())
    ctx.CustomData.Set("user_ip", r.RemoteAddr)

    // Get custom data
    startTime, _ := ctx.CustomData.Get("start_time")
    userIP, _ := ctx.CustomData.Get("user_ip")

    // Use the data...
})
```

## Complete Example

Here's a comprehensive example showing most features:

```go
package main

import (
    "fmt"
    "log"
    "net/http"
    "time"

    api "github.com/your-username/go-restapi"
)

// Define permissions
const (
    PermissionViewUsers   api.Permission = 1
    PermissionEditUsers   api.Permission = 2
    PermissionDeleteUsers api.Permission = 3
    PermissionAdmin       api.Permission = 10
)

// User represents a user in the system
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func main() {
    // Configure CORS
    corsConfig := &api.CORSConfig{
        AllowedOrigins:   []string{"https://myapp.com", "http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
        MaxAge:           3600,
    }

    // Create router
    router := &api.Router{
        BasePath:   "/api/v1",
        CORSConfig: corsConfig,
    }

    // Set up authentication middleware
    router.AuthorizationMiddleware = func(context *api.RouteContext, handler http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")

            // Mock token validation
            if token == "Bearer valid-token" {
                context.SetUserId("user-123")
                context.Params.Set("user_role", "admin")
                handler.ServeHTTP(w, r)
            } else {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
            }
        })
    }

    // Set up permission middleware
    router.PermissionMiddleware = func(context *api.RouteContext, handler http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Mock user permissions (in real app, fetch from database)
            userPermissions := []api.Permission{PermissionViewUsers, PermissionEditUsers, PermissionAdmin}

            if context.HasRequiredPermissions(userPermissions) {
                handler.ServeHTTP(w, r)
            } else {
                http.Error(w, "Forbidden", http.StatusForbidden)
            }
        })
    }

    // Public routes
    router.HandleFunc("GET", "/health", func(w http.ResponseWriter, r *http.Request, ctx *api.RouteContext) {
        api.WriteJSON(w, map[string]string{"status": "ok", "time": time.Now().Format(time.RFC3339)})
    })

    router.HandleFunc("GET", "/users/:id", func(w http.ResponseWriter, r *http.Request, ctx *api.RouteContext) {
        id, err := ctx.Params.Get("id")
        if err != nil {
            http.Error(w, "Invalid user ID", http.StatusBadRequest)
            return
        }

        user := User{ID: 1, Name: "John Doe", Email: "john@example.com"}
        api.WriteJSON(w, user)
    })

    // Protected routes
    router.HandleProtectedFunc("GET", "/admin/users",
        []api.Permission{PermissionViewUsers},
        func(w http.ResponseWriter, r *http.Request, ctx *api.RouteContext) {
            users := []User{
                {ID: 1, Name: "John Doe", Email: "john@example.com"},
                {ID: 2, Name: "Jane Smith", Email: "jane@example.com"},
            }
            api.WriteJSON(w, users)
        })

    router.HandleProtectedFunc("DELETE", "/admin/users/:id",
        []api.Permission{PermissionDeleteUsers, PermissionAdmin},
        func(w http.ResponseWriter, r *http.Request, ctx *api.RouteContext) {
            id, _ := ctx.Params.Get("id")
            userId, _ := ctx.GetUserId()

            // Log the deletion
            fmt.Printf("User %s deleted user %s\\n", userId, id)

            api.WriteJSON(w, map[string]string{"message": "User deleted", "id": id})
        })

    // Set up logging
    api.SetRedactedHeaderNames([]string{"Authorization"})
    logFunc := func(entry api.HttpLogEntry) {
        fmt.Printf("[%s] %s %s %d (trace: %s)\\n",
            time.Now().Format("15:04:05"),
            entry.Method,
            entry.Path,
            entry.Status,
            entry.TraceID)
    }

    // Apply middleware
    loggedRouter := api.LoggingRouter(router, logFunc)
    tracedRouter := api.TracingRouter(loggedRouter)

    // Start server
    fmt.Println("Server starting on :8080")
    fmt.Println("Try these endpoints:")
    fmt.Println("  GET  http://localhost:8080/api/v1/health")
    fmt.Println("  GET  http://localhost:8080/api/v1/users/123")
    fmt.Println("  GET  http://localhost:8080/api/v1/admin/users (requires: Authorization: Bearer valid-token)")

    log.Fatal(http.ListenAndServe(":8080", tracedRouter))
}
```

## API Reference

### Types

#### Router

```go
type Router struct {
    BasePath                string
    Routes                  []Route
    AuthorizationMiddleware func(context *RouteContext, handler http.Handler) http.Handler
    PermissionMiddleware    func(context *RouteContext, handler http.Handler) http.Handler
    CORSConfig              *CORSConfig
}
```

#### RouteContext

```go
type RouteContext struct {
    Params     *RouteParams  // URL parameters
    CustomData *CustomData   // Custom request-scoped data
}

// Methods
func (rc *RouteContext) GetUserId() (string, error)
func (rc *RouteContext) SetUserId(userId string)
func (rc *RouteContext) HasRequiredPermissions(userPermissions []Permission) bool
func (rc *RouteContext) GetRequiredPermissions() ([]Permission, error)
```

#### CORSConfig

```go
type CORSConfig struct {
    AllowedOrigins   []string
    AllowedMethods   []string
    AllowedHeaders   []string
    AllowCredentials bool
    MaxAge           int
}
```

### Functions

#### Router Methods

- `HandleFunc(method, path string, handler RouteHandlerFunc)`
- `HandleProtectedFunc(method, path string, permissions []Permission, handler RouteHandlerFunc)`

#### JSON Utilities

- `WriteJSON(w http.ResponseWriter, data interface{}) error`
- `WriteJSONWithoutTemplate(w http.ResponseWriter, data interface{}) error`
- `ReadJSON(r *http.Request, v interface{}) error`
- `SetJSONResponseFormatter(f func(interface{}) interface{})`

#### Middleware

- `LoggingRouter(next http.Handler, logFunc func(entry HttpLogEntry)) http.Handler`
- `TracingRouter(next http.Handler) http.Handler`
- `SetRedactedHeaderNames(headerNames []string)`

#### Multi-Router

- `NewMultiRouter(basePath string, routers []*Router) (*MultiRouter, error)`

## Best Practices

1. **Define permissions as constants** in your application
2. **Use middleware** for cross-cutting concerns (logging, tracing, auth)
3. **Configure CORS** appropriately for your environment
4. **Handle errors gracefully** in your handlers
5. **Use protected routes** for sensitive operations
6. **Implement proper token validation** in AuthorizationMiddleware
7. **Cache user permissions** to avoid database hits on every request

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
