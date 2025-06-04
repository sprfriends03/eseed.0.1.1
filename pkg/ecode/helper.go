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

// errorMessages maps error codes to their user-friendly messages
var errorMessages = make(map[int]string)

// RegisterError registers a new error code and its associated message
func RegisterError(code int, message string) {
	if _, exists := errorMessages[code]; exists {
		logrus.Warnf("Error code %d already registered, overwriting", code)
	}
	errorMessages[code] = message
}

// GetMessage returns the message associated with an error code
func GetMessage(code int) string {
	if message, ok := errorMessages[code]; ok {
		return message
	}
	return fmt.Sprintf("Unknown error (code: %d)", code)
}

// NewWithContext creates a new cannabis-specific error with context
func NewWithContext(code int, contextStr string) *Error {
	message := GetMessage(code)
	if contextStr != "" {
		message = fmt.Sprintf("%s: %s", message, contextStr)
	}

	err := New(500, message)
	err.ErrCode = fmt.Sprintf("%d", code) // Store code as string in ErrCode
	stackTrace := captureStack()
	err.ErrStack = stackTrace // Store stack in ErrStack
	// Store context in ErrDesc if empty
	if err.ErrDesc == "" {
		err.ErrDesc = contextStr
	}

	return err
}

// Newf creates a new formatted cannabis-specific error
func Newf(code int, format string, args ...interface{}) *Error {
	message := GetMessage(code)
	contextMsg := fmt.Sprintf(format, args...)
	fullMessage := fmt.Sprintf("%s: %s", message, contextMsg)

	err := New(500, fullMessage)
	err.ErrCode = fmt.Sprintf("%d", code)
	stackTrace := captureStack()
	err.ErrStack = stackTrace

	return err
}

// WrapError wraps an existing error with a cannabis-specific error code
func WrapError(code int, origErr error, contextStr string) *Error {
	if origErr == nil {
		return nil
	}

	message := GetMessage(code)
	if contextStr != "" {
		message = fmt.Sprintf("%s: %s", message, contextStr)
	}

	wrappedErr := New(500, fmt.Sprintf("%s: %v", message, origErr))
	wrappedErr.ErrCode = fmt.Sprintf("%d", code)
	wrappedErr.ErrStack = captureStack()
	wrappedErr.ErrDesc = contextStr

	// Add original error info to description
	if wrappedErr.ErrDesc != "" {
		wrappedErr.ErrDesc = fmt.Sprintf("%s (caused by: %v)", wrappedErr.ErrDesc, origErr)
	} else {
		wrappedErr.ErrDesc = fmt.Sprintf("caused by: %v", origErr)
	}

	return wrappedErr
}

// WrapErrorf wraps an existing error with a cannabis-specific error code and formatted context
func WrapErrorf(code int, origErr error, format string, args ...interface{}) *Error {
	if origErr == nil {
		return nil
	}

	message := GetMessage(code)
	contextMsg := fmt.Sprintf(format, args...)
	fullMessage := fmt.Sprintf("%s: %s", message, contextMsg)

	wrappedErr := New(500, fmt.Sprintf("%s: %v", fullMessage, origErr))
	wrappedErr.ErrCode = fmt.Sprintf("%d", code)
	wrappedErr.ErrStack = captureStack()
	wrappedErr.ErrDesc = contextMsg

	// Add original error info to description
	if wrappedErr.ErrDesc != "" {
		wrappedErr.ErrDesc = fmt.Sprintf("%s (caused by: %v)", wrappedErr.ErrDesc, origErr)
	} else {
		wrappedErr.ErrDesc = fmt.Sprintf("caused by: %v", origErr)
	}

	return wrappedErr
}

// IsCode checks if the given error has the specified code
func IsCode(err error, code int) bool {
	if err == nil {
		return false
	}
	if codeErr, ok := err.(*Error); ok {
		return codeErr.ErrCode == fmt.Sprintf("%d", code)
	}
	return false
}

// LogError logs an error with its details to logrus
func LogError(err error) {
	if err == nil {
		return
	}

	if codeErr, ok := err.(*Error); ok {
		fields := logrus.Fields{
			"error_code": codeErr.ErrCode,
		}

		if codeErr.ErrDesc != "" {
			fields["context"] = codeErr.ErrDesc
		}

		if codeErr.ErrStack != "" {
			fields["stack"] = codeErr.ErrStack
		}

		logrus.WithFields(fields).Error(codeErr.Error())
	} else {
		logrus.Error(err)
	}
}

// LogWithFields logs an error with additional fields
func LogWithFields(err error, fields logrus.Fields) {
	if err == nil {
		return
	}

	if codeErr, ok := err.(*Error); ok {
		fields["error_code"] = codeErr.ErrCode

		if codeErr.ErrDesc != "" {
			fields["context"] = codeErr.ErrDesc
		}

		if codeErr.ErrStack != "" {
			fields["stack"] = codeErr.ErrStack
		}

		logrus.WithFields(fields).Error(codeErr.Error())
	} else {
		logrus.WithFields(fields).Error(err)
	}
}

// captureStack captures the current stack trace
func captureStack() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var builder strings.Builder
	builder.WriteString("Stack Trace:\n")

	for {
		frame, more := frames.Next()
		if !strings.Contains(frame.File, "github.com") && !strings.Contains(frame.File, "app/") {
			// Skip standard library frames
			if more {
				continue
			}
			break
		}

		builder.WriteString(fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function))
		if !more {
			break
		}
	}

	return builder.String()
}

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
