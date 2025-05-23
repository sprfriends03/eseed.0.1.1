package ecode

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"app/pkg/trace"
)

// Examples of how to use the error handling framework
// These are just examples and not meant to be executed directly

// ExampleBasicErrorHandling demonstrates basic error creation and handling
func ExampleBasicErrorHandling() error {
	// Basic error creation
	err := BadRequest.Desc(errors.New("invalid parameter: id"))

	// Using WithDesc shorthand
	err = WithDesc(BadRequest, "invalid parameter: id")

	// Return the error to middleware for handling
	// The middleware will log the error and convert to the appropriate response
	return err
}

// validateInput is a simple example function that returns an error
func validateInput(id string) error {
	if id == "" {
		return BadRequest.Desc(errors.New("id is required"))
	}
	return nil
}

// ExampleErrorWithStack demonstrates adding stack information to errors
func ExampleErrorWithStack() error {
	// Get error with stack trace information
	err := WithStack(errors.New("something went wrong"))

	// Or add stack to domain-specific errors
	err = WithStack(PlantNotFound)

	// Log with proper level and return
	return LogError(err)
}

// ExampleErrorWithContext demonstrates adding context information to errors
func ExampleErrorWithContext() error {
	// Create a context with tracing
	ctx := trace.New(context.Background())

	// Add context value for better debugging
	ctx = trace.WithValue(ctx, trace.E{
		K: "plant_id",
		V: "abc123",
	})

	// Wrap error with context information
	err := WithContext(ctx, PlantNotFound)

	// Log with context and return
	return LogErrorWithContext(ctx, err)
}

// Plant is a placeholder type for the example
type Plant struct{}

// ExampleHandlingDatabaseError demonstrates handling database errors
func ExampleHandlingDatabaseError(ctx context.Context, id string) (*Plant, error) {
	// This is just an example function signature
	// db.FindPlantByID would be a real database call
	var plant *Plant

	// Simulating a database error (in real code, this would be a real DB call)
	err := errors.New("simulated database error")

	if err != nil {
		if errors.Is(err, errors.New("no documents")) {
			// Convert to domain-specific error
			return nil, WithContext(ctx, PlantNotFound.Desc(fmt.Errorf("plant not found: %s", id)))
		}
		// Add context to generic database error
		return nil, WithContext(ctx, InternalServerError.Stack(err))
	}

	return plant, nil
}

// Handler is a placeholder type for the example
type Handler struct {
	service struct {
		GetPlant func(ctx context.Context, id string) (*Plant, error)
	}
}

// GinLikeContext simulates a gin.Context for the example
type GinLikeContext interface {
	Request() RequestLike
	Param(name string) string
	Error(err error)
	JSON(code int, obj interface{})
}

// RequestLike simulates an http.Request for the example
type RequestLike interface {
	Context() context.Context
}

// ExampleUsingInHandler demonstrates using errors in HTTP handlers
func (h *Handler) ExampleUsingInHandler(c GinLikeContext) {
	// This is just an example using an interface that mimics Gin context
	ctx := c.Request().Context()
	id := c.Param("id")

	// In a real handler, this would call a real service
	plant, err := h.service.GetPlant(ctx, id)
	if err != nil {
		// Just return the error - middleware handles logging and response
		c.Error(err)
		return
	}

	c.JSON(http.StatusOK, plant)
}

// ExampleConditionalErrorWrapping demonstrates conditional error wrapping
func ExampleConditionalErrorWrapping() error {
	// Simulating a function call that might return an error
	err := errors.New("some error")

	if err != nil {
		// Only wrap if it's a specific type
		return WrapIf(err,
			func(err error) bool {
				// Check if it's a domain-specific error
				_, ok := err.(*Error)
				return !ok
			},
			func(err error) error {
				// Wrap with internal server error
				return InternalServerError.Stack(err)
			},
		)
	}

	return nil
}
