package v1

import (
	"net/http"

	"github.com/carsonmyers/bublar-assignment/errors"
)

// PingData - request/response struct for ping
type PingData struct {
	Say string `json:"say"`
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	var req PingData
	if err := DecodeRequest(w, r, &req); err != nil {
		return
	}

	if req.Say == "ping" {
		pong := PingData{
			Say: "pong",
		}

		FromData(pong).Write(w)
	} else {
		FromError(errors.EInvalidRequest.NewErrorf("Expected \"ping\", received \"%s\"", req.Say).WithContext("say")).Write(w)
	}
}
