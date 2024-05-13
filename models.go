package restapi

import (
	"encoding/json"
	"net/http"
	"time"
)

type Response struct {
	Timestamp int64       `json:"timestamp"`
	Success   bool        `json:"success"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
}

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
	json.NewEncoder(sw).Encode(data)
}
