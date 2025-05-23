# Error Handling Framework

This package provides a comprehensive error handling framework for the Seed eG Platform. It includes:

1. Standardized error types with HTTP status codes
2. Cannabis club-specific error codes
3. Error context enrichment with stack traces and request context
4. Structured error logging with appropriate severity levels
5. Helper functions for consistent error handling across the application

## Key Components

- **index.go**: Core error types and base error codes
- **cannabis.go**: Domain-specific error codes for cannabis club platform
- **helper.go**: Utility functions for error handling, context, and logging
- **examples.go**: Example patterns for using the error framework

## Usage Examples

### Basic Error Handling

```go
// Create a new error with description
err := BadRequest.Desc(errors.New("invalid parameter: id"))

// Using WithDesc shorthand
err = WithDesc(BadRequest, "invalid parameter: id")
```

### Adding Stack Traces

```go
// Add stack trace information to an error
err := WithStack(errors.New("something went wrong"))

// Or add stack to domain-specific errors
err = WithStack(PlantNotFound)
```

### Context-Aware Errors

```go
// Add context information to errors
err := WithContext(ctx, PlantNotFound)

// Log with context information
LogErrorWithContext(ctx, err)
```

### Error Handling in Route Handlers

```go
func (h *Handler) GetPlant(c *gin.Context) {
    ctx := c.Request.Context()
    id := c.Param("id")
    
    plant, err := h.service.GetPlant(ctx, id)
    if err != nil {
        // Just return the error - middleware handles logging and response
        c.Error(err)
        return
    }
    
    c.JSON(http.StatusOK, plant)
}
```

## Error Categories

The framework includes specialized error codes for different functional areas:

- **KYC-related errors**: Verification requirements, document validation
- **Membership-related errors**: Expiration, limits, requirements
- **Plant slot errors**: Availability, allocation
- **Plant errors**: Status, care, image handling
- **Harvest errors**: Readiness, collection status
- **NFT-related errors**: Minting, verification
- **Payment errors**: Processing, requirements
- **Validation errors**: Data format, required fields

## Best Practices

1. **Use domain-specific errors** when possible for more precise error handling
2. **Add context information** to errors using `WithContext` for better debugging
3. **Include stack traces** with `WithStack` for internal server errors
4. **Log errors appropriately** using `LogError` or `LogErrorWithContext`
5. **Return errors directly** in route handlers and let middleware handle logging and response
6. **Wrap external errors** with appropriate domain errors using error helpers

## Integration with Middleware

The error handling framework works seamlessly with the existing middleware chain:

```go
// From route/index.go
func (s middleware) Error() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()
        if len(c.Errors) > 0 {
            s.ErrorFunc(c, c.Errors.Last().Err)
        }
    }
}

func (s middleware) ErrorFunc(c *gin.Context, err any) {
    switch e := err.(type) {
    case *ecode.Error:
        c.JSON(e.Status, e)
    default:
        err := ecode.InternalServerError.Stack(fmt.Errorf("%v", e))
        c.JSON(err.Status, err)
    }
}
``` 