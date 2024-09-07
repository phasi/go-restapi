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

// WriteJSON writes a JSON response to the ResponseWriter
func WriteJSON(w http.ResponseWriter, data interface{}) {
	sw := &statusWriter{ResponseWriter: w}
	sw.Header().Set("Content-Type", "application/json")
	if sw.status == 0 {
		if data == nil {
			sw.WriteHeader(http.StatusNoContent)
			return
		} else {
			sw.WriteHeader(http.StatusOK)
		}
	}
	if data == nil {
		data = Response{
			Timestamp: time.Now().Unix(),
			Success:   true,
			Message:   "Success",
			Data:      nil,
		}
	} else {
		data = Response{
			Timestamp: time.Now().Unix(),
			Success:   true,
			Message:   "Success",
			Data:      data,
		}
	}
	json.NewEncoder(sw).Encode(data) // TODO: handle error
}

// ReadJSON reads a JSON request from the Request and decodes it into the provided interface
func ReadJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}
