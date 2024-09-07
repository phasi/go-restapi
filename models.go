package restapi

import (
	"encoding/json"
	"net/http"
	"time"
)

type Response struct {
	Timestamp int64       `json:"timestamp"`
	Data      interface{} `json:"data"`
}

func getDefaultJSONResponse(data interface{}) interface{} {
	if data == nil {
		return Response{
			Timestamp: time.Now().Unix(),
			Data:      nil,
		}
	} else {
		return Response{
			Timestamp: time.Now().Unix(),
			Data:      data,
		}
	}
}

var jsonResponseFormatter func(interface{}) interface{} = getDefaultJSONResponse

func SetJSONResponseFormatter(f func(interface{}) interface{}) {
	jsonResponseFormatter = f
}

func writeJSON(w http.ResponseWriter, data interface{}, usesTemplate bool) error {
	sw := &statusWriter{ResponseWriter: w}
	sw.Header().Set("Content-Type", "application/json")
	if sw.status == 0 {
		if data == nil {
			sw.WriteHeader(http.StatusNoContent)
			return nil
		} else {
			sw.WriteHeader(http.StatusOK)
		}
	}
	if usesTemplate {
		data = jsonResponseFormatter(data)
	}
	return json.NewEncoder(sw).Encode(data)
}

// WriteJSON writes a JSON response to the ResponseWriter
func WriteJSON(w http.ResponseWriter, data interface{}) error {
	return writeJSON(w, data, true)
}

func WriteJSONWithoutTemplate(w http.ResponseWriter, data interface{}) error {
	return writeJSON(w, data, false)
}

// ReadJSON reads a JSON request from the Request and decodes it into the provided interface
func ReadJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
