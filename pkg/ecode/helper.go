package ecode

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"

	"app/pkg/trace"
	"context"

	"github.com/sirupsen/logrus"
)

// WithContext wraps an error with context information from the trace context
func WithContext(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	// If it's already our error type, just add trace info
	if e, ok := err.(*Error); ok {
		return addTraceInfo(ctx, e)
	}

	// Otherwise wrap with InternalServerError
	return addTraceInfo(ctx, InternalServerError.Stack(err))
}

// WithStack adds stack trace information to the error
func WithStack(err error) error {
	if err == nil {
		return nil
	}

	// Get caller information
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		return err
	}

	// Extract function name
	fn := runtime.FuncForPC(pc)
	funcName := fn.Name()
	parts := strings.Split(funcName, ".")
	callerInfo := parts[len(parts)-1]

	// If it's already our error type, add stack info
	if e, ok := err.(*Error); ok {
		stackInfo := fmt.Sprintf("%s at %s:%d", callerInfo, file, line)
		if e.ErrStack == "" {
			e.ErrStack = stackInfo
		} else {
			e.ErrStack = fmt.Sprintf("%s <- %s", e.ErrStack, stackInfo)
		}
		return e
	}

	// Otherwise wrap with InternalServerError
	return InternalServerError.Stack(fmt.Errorf("%v (at %s:%d in %s)", err, file, line, callerInfo))
}

// LogError logs the error with appropriate level and returns the error for chaining
func LogError(err error) error {
	if err == nil {
		return nil
	}

	// Format error for logging
	errorJSON := formatErrorForLog(err)

	// Log based on HTTP status if it's our error type
	if e, ok := err.(*Error); ok {
		switch {
		case e.Status >= 500:
			logrus.WithField("error_details", errorJSON).Error(e.Error())
		case e.Status >= 400:
			logrus.WithField("error_details", errorJSON).Warn(e.Error())
		default:
			logrus.WithField("error_details", errorJSON).Info(e.Error())
		}
		return e
	}

	// Generic error
	logrus.WithField("error_details", errorJSON).Error(err.Error())
	return err
}

// LogErrorWithContext logs the error with context and returns the error for chaining
func LogErrorWithContext(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	// Add trace ID to error context
	errorData := map[string]interface{}{
		"trace_id": trace.ID(ctx),
		"time":     time.Now().Format(time.RFC3339),
	}

	// Format error for logging
	errorJSON := formatErrorWithContextForLog(ctx, err, errorData)

	// Log based on HTTP status if it's our error type
	if e, ok := err.(*Error); ok {
		switch {
		case e.Status >= 500:
			logrus.WithField("error_context", errorJSON).Error(e.Error())
		case e.Status >= 400:
			logrus.WithField("error_context", errorJSON).Warn(e.Error())
		default:
			logrus.WithField("error_context", errorJSON).Info(e.Error())
		}
		return e
	}

	// Generic error
	logrus.WithField("error_context", errorJSON).Error(err.Error())
	return err
}

// WrapIf wraps an error only if the predicate function returns true
func WrapIf(err error, predicate func(error) bool, wrapper func(error) error) error {
	if err == nil {
		return nil
	}

	if predicate(err) {
		return wrapper(err)
	}

	return err
}

// Internal helper functions

// addTraceInfo adds trace context information to an error
func addTraceInfo(ctx context.Context, err *Error) *Error {
	traceID := trace.ID(ctx)
	if traceID != "" && err.ErrStack != "" {
		// Only append trace ID if we have one and there's already stack info
		err.ErrStack = fmt.Sprintf("%s (trace:%s)", err.ErrStack, traceID)
	} else if traceID != "" {
		err.ErrStack = fmt.Sprintf("trace:%s", traceID)
	}
	return err
}

// formatErrorForLog formats an error for logging
func formatErrorForLog(err error) string {
	errorData := map[string]interface{}{
		"time": time.Now().Format(time.RFC3339),
	}

	if e, ok := err.(*Error); ok {
		errorData["code"] = e.ErrCode
		errorData["description"] = e.ErrDesc
		errorData["stack"] = e.ErrStack
		errorData["status"] = e.Status
	} else {
		errorData["error"] = err.Error()
	}

	data, _ := json.Marshal(errorData)
	return string(data)
}

// formatErrorWithContextForLog formats an error with context data for logging
func formatErrorWithContextForLog(ctx context.Context, err error, contextData map[string]interface{}) string {
	// Get trace values
	traceValues := trace.Value(ctx)
	if len(traceValues) > 0 {
		traceData := make(map[string]interface{})
		for _, e := range traceValues {
			traceData[e.K] = e.V
		}
		contextData["trace_values"] = traceData
	}

	if e, ok := err.(*Error); ok {
		contextData["code"] = e.ErrCode
		contextData["description"] = e.ErrDesc
		contextData["stack"] = e.ErrStack
		contextData["status"] = e.Status
	} else {
		contextData["error"] = err.Error()
	}

	data, _ := json.Marshal(contextData)
	return string(data)
}
