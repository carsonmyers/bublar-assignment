package v1

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/carsonmyers/bublar-assignment/errors"
	"go.uber.org/zap"
)

// DecodeRequest unmarshals a JSON request body into a supplied interface
func DecodeRequest(w http.ResponseWriter, r *http.Request, target interface{}) error {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Error("Failed to read request body", zap.Error(err))
		FromError(errors.EInternal.NewError("Error reading request")).Write(w)
		return err
	}

	err = json.Unmarshal(data, target)
	if err != nil {
		log.Error("Failed to unmarshal request body", zap.Error(err))
		FromError(errors.EInternal.NewError("Error reading request")).Write(w)
		return err
	}

	return nil
}
