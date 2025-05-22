package api

import (
	"app/pkg/ecode"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

type Request[T any] struct {
	r   *http.Request
	err error
}

type Response[T any] struct {
	Data   T
	Body   []byte
	Error  *ecode.Error
	Status int
}

func New[T any](r *http.Request, errs ...error) *Request[T] {
	if len(errs) > 0 {
		return &Request[T]{r, errs[0]}
	}
	return &Request[T]{r, nil}
}

func (s *Request[T]) buildResponse(data T, body []byte, err *ecode.Error) *Response[T] {
	if err != nil {
		return &Response[T]{Data: data, Body: body, Error: err, Status: err.Status}
	}
	return &Response[T]{Data: data, Body: body, Error: nil, Status: http.StatusOK}
}

func (s *Request[T]) SetHeader(key, value string) *Request[T] {
	s.r.Header.Set(key, value)
	return s
}

func (s *Request[T]) SetContentType(contentType string) *Request[T] {
	s.r.Header.Set("Content-Type", contentType)
	return s
}

func (s *Request[T]) SetAuthorization(authorization string) *Request[T] {
	s.r.Header.Set("Authorization", authorization)
	return s
}

func (s *Request[T]) SetBearerAuth(bearer string) *Request[T] {
	s.r.Header.Set("Authorization", "Bearer "+bearer)
	return s
}

func (s *Request[T]) SetBasicAuth(username, password string) *Request[T] {
	s.r.SetBasicAuth(username, password)
	return s
}

func (s *Request[T]) SetForm(key, value string) *Request[T] {
	s.r.Form.Set(key, value)
	return s
}

func (s *Request[T]) SetBody(body []byte) *Request[T] {
	s.r.Body = io.NopCloser(bytes.NewBuffer(body))
	return s
}

func (s *Request[T]) Call() *Response[T] {
	var data T
	var body []byte

	if s.err != nil {
		return s.buildResponse(data, body, ecode.InternalServerError.Desc(s.err))
	}

	if len(s.r.Header.Get("Content-Type")) == 0 {
		s.SetContentType("application/json")
	}

	resp, err := http.DefaultClient.Do(s.r)
	if err != nil {
		return s.buildResponse(data, body, ecode.InternalServerError.Desc(err))
	}
	defer resp.Body.Close()

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return s.buildResponse(data, body, ecode.InternalServerError.Desc(err))
	}

	if resp.StatusCode >= http.StatusBadRequest {
		var e *ecode.Error
		if err = json.Unmarshal(body, &e); err != nil {
			return s.buildResponse(data, body, ecode.InternalServerError.Desc(err))
		}
		e.Status = resp.StatusCode
		if len(e.ErrCode) == 0 {
			e.ErrCode = strings.ToLower(strings.ReplaceAll(resp.Status, " ", "_"))
			e.ErrDesc = string(body)
		}
		return s.buildResponse(data, body, e)
	}

	if err = json.Unmarshal(body, &data); err != nil {
		return s.buildResponse(data, body, ecode.InternalServerError.Desc(err))
	}

	return s.buildResponse(data, body, nil)
}
