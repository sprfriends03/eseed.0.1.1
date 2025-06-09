package ecode

import (
	"errors"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
)

// Common error codes
var (
	InternalServerError     = New(http.StatusInternalServerError, "internal_server_error")
	TooManyRequests         = New(http.StatusTooManyRequests, "too_many_requests")
	UpgradeRequired         = New(http.StatusUpgradeRequired, "upgrade_required")
	Unauthorized            = New(http.StatusUnauthorized, "unauthorized")
	Forbidden               = New(http.StatusForbidden, "forbidden")
	BadRequest              = New(http.StatusBadRequest, "bad_request")
	FileNotFound            = New(http.StatusBadRequest, "file_not_found")
	OldPasswordIncorrect    = New(http.StatusBadRequest, "old_password_incorrect")
	UserOrPasswordIncorrect = New(http.StatusBadRequest, "user_or_password_incorrect")
	InvalidToken            = New(http.StatusBadRequest, "invalid_token")
	InvalidSignature        = New(http.StatusBadRequest, "invalid_signature")
	InvalidPermission       = New(http.StatusBadRequest, "invalid_permission")
	ApiNotFound             = New(http.StatusNotFound, "api_not_found")
	ClientNotFound          = New(http.StatusNotFound, "client_key_not_found")
	ClientConflict          = New(http.StatusConflict, "client_key_conflict")
	TenantNotFound          = New(http.StatusNotFound, "tenant_not_found")
	TenantConflict          = New(http.StatusConflict, "tenant_conflict")
	RoleNotFound            = New(http.StatusNotFound, "role_not_found")
	RoleConflict            = New(http.StatusConflict, "role_conflict")
	UserNotFound            = New(http.StatusNotFound, "user_not_found")
	UserConflict            = New(http.StatusConflict, "user_conflict")

	// Membership specific errors (only new ones not in cannabis.go)
	MembershipNotFound      = New(http.StatusNotFound, "membership_not_found")
	MembershipConflict      = New(http.StatusConflict, "membership_conflict")
	InvalidMembershipType   = New(http.StatusBadRequest, "invalid_membership_type")
	MembershipAlreadyActive = New(http.StatusConflict, "membership_already_active")
	InsufficientSlots       = New(http.StatusConflict, "insufficient_slots")
	MembershipRenewalFailed = New(http.StatusBadRequest, "membership_renewal_failed")
)

// Error represents an API error with HTTP status, error code, and descriptive text.
// Additionally stores stack information for logging purposes.
type Error struct {
	Status   int    `json:"-"`                 // HTTP status code
	ErrCode  string `json:"error"`             // Machine-readable error code
	ErrDesc  string `json:"error_description"` // Human-readable error description
	ErrStack string `json:"-"`                 // Stack trace information (not sent to client)
}

// New creates a new Error with the given status code and error code.
// The error description is initially empty and can be set with Desc().
func New(status int, code string) *Error {
	return &Error{Status: status, ErrCode: code}
}

// Desc sets the error description and returns the error for chaining.
// This provides human-readable context for the error.
func (s *Error) Desc(err error) *Error {
	s.ErrDesc = err.Error()
	return s
}

// Stack sets the error stack trace information and returns the error for chaining.
// This provides technical context for debugging but is not exposed to clients.
// Also handles MongoDB duplicate key errors by converting to Conflict status.
func (s *Error) Stack(err error) *Error {
	// Handle MongoDB duplicate key errors specially
	if s.Status == http.StatusConflict && !mongo.IsDuplicateKeyError(err) {
		s = InternalServerError
	}
	s.ErrStack = err.Error()
	return s
}

// Error implements the error interface and returns a string representation
// of the error, including code, description, and stack trace if available.
func (s *Error) Error() string {
	errs := make([]string, 0)
	errs = append(errs, s.ErrCode)
	if len(s.ErrDesc) > 0 {
		errs = append(errs, s.ErrDesc)
	}
	if len(s.ErrStack) > 0 {
		errs = append(errs, s.ErrStack)
	}
	return strings.Join(errs, " | ")
}

// WithDesc is a shorthand for creating an error with a description in one line.
func WithDesc(code *Error, desc string) *Error {
	return code.Desc(errors.New(desc))
}
