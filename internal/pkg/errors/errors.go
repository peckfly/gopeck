package errors

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"sync"
)

// Define alias
var (
	WithStack = errors.WithStack
	Wrap      = errors.Wrap
	Wrapf     = errors.Wrapf
	Is        = errors.Is
	Errorf    = errors.Errorf
)

const (
	DefaultBadRequestID            = "bad_request"
	DefaultUnauthorizedID          = "unauthorized"
	DefaultForbiddenID             = "forbidden"
	DefaultNotFoundID              = "not_found"
	DefaultMethodNotAllowedID      = "method_not_allowed"
	DefaultTooManyRequestsID       = "too_many_requests"
	DefaultRequestEntityTooLargeID = "request_entity_too_large"
	DefaultInternalServerErrorID   = "internal_server_error"
	DefaultConflictID              = "conflict"
	DefaultRequestTimeoutID        = "request_timeout"
)

// Error Customize the error structure for implementation errors.Error interface
type Error struct {
	ID     string `json:"id,omitempty"`
	Code   int32  `json:"code,omitempty"`
	Detail string `json:"detail,omitempty"`
	Status string `json:"status,omitempty"`
}

func (e *Error) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}

// New generates a custom error.
func New(id, detail string, code int32) error {
	return &Error{
		ID:     id,
		Code:   code,
		Detail: detail,
		Status: http.StatusText(int(code)),
	}
}

// Parse tries to parse a JSON string into an error. If that
// fails, it will set the given string as the error detail.
func Parse(err string) *Error {
	e := new(Error)
	err1 := json.Unmarshal([]byte(err), e)
	if err1 != nil {
		e.Detail = err
	}
	return e
}

// BadRequest generates a 400 error.
func BadRequest(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultBadRequestID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusBadRequest,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusBadRequest),
	}
}

// Unauthorized generates a 401 error.
func Unauthorized(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultUnauthorizedID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusUnauthorized,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusUnauthorized),
	}
}

// Forbidden generates a 403 error.
func Forbidden(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultForbiddenID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusForbidden,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusForbidden),
	}
}

// NotFound generates a 404 error.
func NotFound(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultNotFoundID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusNotFound,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusNotFound),
	}
}

// MethodNotAllowed generates a 405 error.
func MethodNotAllowed(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultMethodNotAllowedID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusMethodNotAllowed,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusMethodNotAllowed),
	}
}

// TooManyRequests generates a 429 error.
func TooManyRequests(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultTooManyRequestsID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusTooManyRequests,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusTooManyRequests),
	}
}

// Timeout generates a 408 error.
func Timeout(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultRequestTimeoutID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusRequestTimeout,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusRequestTimeout),
	}
}

// Conflict generates a 409 error.
func Conflict(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultConflictID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusConflict,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusConflict),
	}
}

// RequestEntityTooLarge generates a 413 error.
func RequestEntityTooLarge(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultRequestEntityTooLargeID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusRequestEntityTooLarge,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusRequestEntityTooLarge),
	}
}

// InternalServerError generates a 500 error.
func InternalServerError(id, format string, a ...interface{}) error {
	if id == "" {
		id = DefaultInternalServerErrorID
	}
	return &Error{
		ID:     id,
		Code:   http.StatusInternalServerError,
		Detail: fmt.Sprintf(format, a...),
		Status: http.StatusText(http.StatusInternalServerError),
	}
}

// Equal tries to compare errors
func Equal(err1 error, err2 error) bool {
	var err11 *Error
	ok1 := errors.As(err1, &err11)
	var err22 *Error
	ok2 := errors.As(err2, &err22)

	if ok1 != ok2 {
		return false
	}

	if !ok1 {
		return errors.Is(err2, err1)
	}

	if err11.Code != err22.Code {
		return false
	}

	return true
}

// FromError try to convert go error to *Error
func FromError(err error) *Error {
	if err == nil {
		return nil
	}
	var err1 *Error
	if errors.As(err, &err1) && err1 != nil {
		return err1
	}

	return Parse(err.Error())
}

// As finds the first error in err's chain that matches *Error
func As(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	var errs *Error
	if errors.As(err, &errs) {
		return errs, true
	}
	return nil, false
}

type MultiError struct {
	lock   *sync.Mutex
	Errors []error
}

func NewMultiError() *MultiError {
	return &MultiError{
		lock:   &sync.Mutex{},
		Errors: make([]error, 0),
	}
}

func (e *MultiError) Append(err error) {
	e.Errors = append(e.Errors, err)
}

func (e *MultiError) AppendWithLock(err error) {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.Append(err)
}

func (e *MultiError) HasErrors() bool {
	return len(e.Errors) > 0
}

func (e *MultiError) Error() string {
	b, _ := json.Marshal(e)
	return string(b)
}
