package types

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func marshalResponse(r Response) string {
	body, err := json.Marshal(r)
	if err != nil {
		return `{"success":false,"message":"internal error"}`
	}
	return string(body)
}

func ErrorResponse(message string) string {
	return marshalResponse(Response{Success: false, Message: message})
}

func SuccessResponse(message string) string {
	return marshalResponse(Response{Success: true, Message: message})
}

func WriteJSON(w http.ResponseWriter, status int, r Response) {
	body, err := json.Marshal(r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"success":false,"message":"internal error"}`))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, Response{Success: false, Message: message})
}

func WriteSuccess(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, Response{Success: true, Message: message})
}
