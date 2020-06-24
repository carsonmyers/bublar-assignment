package v1

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/carsonmyers/bublar-assignment/errors"
	"go.uber.org/zap"
)

// ResponseStatus indicates whether the request succeeded (fully or partially) or failed
type ResponseStatus string

const (
	// StatusOK indicates that a request succeeded
	StatusOK ResponseStatus = "ok"

	// StatusWarn indicates that a request succeeded overall, but that there were problems
	StatusWarn = "warn"

	// StatusError indicates that a request failed
	StatusError = "error"
)

// Response is the overall structure of an API response
type Response struct {
	Status     ResponseStatus  `json:"status"`
	Problems   []*errors.Error `json:"problems"`
	Data       interface{}     `json:"data"`
	statusCode int
}

// NewResponse start a new API response
func NewResponse() *Response {
	return FromData(nil)
}

// FromError createse a response from an error and a code string
func FromError(err *errors.Error) *Response {
	problems := make([]*errors.Error, 1)
	problems[0] = err

	return &Response{
		Status:   StatusError,
		Problems: problems,
		Data:     nil,
	}
}

// FromData creates a successful response from a data struct
func FromData(data interface{}) *Response {
	return &Response{
		Status:   StatusOK,
		Problems: make([]*errors.Error, 0),
		Data:     data,
	}
}

// AddError adds a problem to the response, downgrading its status if it was earlier successful
func (r *Response) AddError(err *errors.Error) *Response {
	r.Problems = append(r.Problems, err)

	if r.Data != nil {
		r.Status = StatusWarn
	} else {
		r.Status = StatusError
	}

	return r
}

// SetData sets the payload data for the response, upgrading its status if it was earlier erroneous
func (r *Response) SetData(data interface{}) *Response {
	r.Data = data
	if r.Status == StatusError {
		r.Status = StatusWarn
	}

	return r
}

// SetStatusCode specifies the status code of the eventual response. If one has
// been set already, the higher (more specific/erroneous, in general) code applies.
func (r *Response) SetStatusCode(statusCode int) *Response {
	if statusCode > r.statusCode {
		r.statusCode = statusCode
	}

	return r
}

const serializeError = `{
	"status": "error",
	"problems": [{
		"ctx": "",
		"errorType": "E_INTERNAL",
		"message": "Error generating response",
	}],
	"data": null
}`

func (r *Response) serialize() []byte {
	data, err := json.Marshal(r)
	if err != nil {
		log.Error("Failed to serialize response", zap.Error(err))
		r.SetStatusCode(http.StatusInternalServerError)
		return []byte(serializeError)
	}

	return data
}

// Write sends the response to a responseWriter
func (r *Response) Write(w http.ResponseWriter) error {
	statusCode := http.StatusOK
	if r.statusCode == 0 {
		for _, err := range r.Problems {
			sc := err.StatusCode()
			if sc > statusCode {
				statusCode = sc
			}
		}
	} else {
		statusCode = r.statusCode
	}

	data := r.serialize()
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, writeErr := w.Write(data)

	if writeErr != nil {
		log.Error("Failed to write response", zap.Error(writeErr))
		return writeErr
	}

	if len(r.Problems) > 0 {
		log.Warn("Problems with request", zap.Int("status", statusCode))
		for i, problem := range r.Problems {
			log.Warn(fmt.Sprintf("\t(%d) %s", i, problem.Error()))
		}
	}

	return nil
}
