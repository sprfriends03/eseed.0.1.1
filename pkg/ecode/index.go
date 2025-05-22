package ecode

import (
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/mongo"
)

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
)

type Error struct {
	Status   int    `json:"-"`
	ErrCode  string `json:"error"`
	ErrDesc  string `json:"error_description"`
	ErrStack string `json:"-"`
}

func New(status int, code string) *Error {
	return &Error{Status: status, ErrCode: code}
}

func (s *Error) Desc(err error) *Error {
	s.ErrDesc = err.Error()
	return s
}

func (s *Error) Stack(err error) *Error {
	if s.Status == http.StatusConflict && !mongo.IsDuplicateKeyError(err) {
		s = InternalServerError
	}
	s.ErrStack = err.Error()
	return s
}

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
