package errors

import (
	"fmt"
	"net/http"
	"strings"
)

// Kind - type alias for types of errors
type Kind string

const (
	// EInternal - internal error that shouldn't be exposed to a frontend
	EInternal = Kind("internal error")

	// EDatabaseConnection - error connecting to a database
	EDatabaseConnection = Kind("failed to connect to database")

	// EDatabase - database operation error
	EDatabase = Kind("database error")

	// ERPCConnection - error connecting to rpc service
	ERPCConnection = Kind("failed to connect to rpc service")

	// ERPC - error originating from a remote procedure call
	ERPC = Kind("remote procedure error")

	// EAuth - error authenticating request
	EAuth = Kind("authentication error")

	// EInvalidRequest - invalid API request
	EInvalidRequest = Kind("invalid request")

	// ENotFound - resource was not found
	ENotFound = Kind("resource not found")

	// ENotImplemented - function is not implemented
	ENotImplemented = Kind("not implemented")

	// EForbidden - user may not access a resource or function
	EForbidden = Kind("forbidden")

	// EDuplicateUser - a username is already taken
	EDuplicateUser = Kind("user already exists")

	// ENotInLocation - the player is not in a location
	ENotInLocation = Kind("not in a location")

	// EDuplicateLocation - a location name is already taken
	EDuplicateLocation = Kind("location already exists")

	// EUnknown - an unknown error occurred
	EUnknown = Kind("unknown error")
)

// New - create a new error
func (ek Kind) New() Error {
	return Error{
		Kind:    ek,
		Message: string(ek),
		Inner:   nil,
	}
}

// NewErrorf - create an error from a kind with a formatted-string inner-error
func (ek Kind) NewErrorf(format string, values ...interface{}) *Error {
	return ek.NewError(fmt.Sprintf(format, values...))
}

// NewError - create an error from a kind with an inner-error
func (ek Kind) NewError(contents interface{}) *Error {
	var msg string
	switch contents.(type) {
	case string:
		msg = contents.(string)
	case error:
		msg = contents.(error).Error()
	case Error:
		msg = contents.(Error).Message
	default:
		msg = fmt.Sprintf("%v", msg)
	}

	return &Error{
		Kind:    ek,
		Message: msg,
		Inner:   nil,
	}
}

// Error - generalized error structure for internal and external use
type Error struct {
	Ctx     string      `json:"ctx"`
	Kind    Kind        `json:"kind"`
	Message string      `json:"message"`
	Inner   interface{} `json:"-"`
}

func (e *Error) Error() string {
	var sb strings.Builder
	sb.WriteString(e.Message)

	var target interface{} = e.Inner
	for {
		if target == nil {
			break
		}

		sb.WriteString(": ")

		switch target.(type) {
		case string:
			sb.WriteString(target.(string))
			target = nil
		case error:
			sb.WriteString(target.(error).Error())
			target = nil
		case *Error:
			sb.WriteString(target.(*Error).Error())
			target = target.(Error).Inner
		default:
			panic("Invalid error type")
		}
	}

	return sb.String()
}

// WithContext - add a context (like and input parameter) to an error
func (e *Error) WithContext(ctx string) *Error {
	e.Ctx = ctx
	return e
}

// Wrap an inner error with this one
func (e *Error) Wrap(contents interface{}) *Error {
	outer := e
	inner := e.Inner
	for {
		if inner == nil {
			outer.Inner = contents
			return e
		}

		switch inner.(type) {
		case *Error:
			outer = inner.(*Error)
			inner = outer.Inner
		default:
			newInner := outer.Kind.NewError(inner)
			newInner.Inner = contents
			outer.Inner = newInner
			return e
		}
	}
}

// StatusCode - derive an HTTP status code from an error kind
func (e Error) StatusCode() int {
	switch e.Kind {
	case EInternal, EDatabaseConnection, EDatabase:
		return http.StatusInternalServerError
	case EInvalidRequest:
		return http.StatusBadRequest
	case ENotFound:
		return http.StatusNotFound
	case ENotImplemented:
		return http.StatusNotImplemented
	case EForbidden:
		return http.StatusForbidden
	case EDuplicateUser, EDuplicateLocation:
		return http.StatusBadRequest
	case ENotInLocation:
		return http.StatusBadRequest
	case EUnknown:
		return http.StatusInternalServerError
	}

	return http.StatusInternalServerError
}
