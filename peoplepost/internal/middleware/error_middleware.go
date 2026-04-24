package middleware

import (
	"encoding/json"
	"net/http"
	"os"
)

type AppError struct {
	StatusCode  int    `json:"-"`
	Message     string `json:"message"`
	Status      string `json:"status"`
	IsOperational bool `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func NewError(statusCode int, message string) *AppError {
	status := "error"
	if statusCode >= 400 && statusCode < 500 {
		status = "fail"
	}
	return &AppError{
		StatusCode:  statusCode,
		Message:     message,
		Status:      status,
		IsOperational: true,
	}
}

func HandleError(w http.ResponseWriter, err error) {
	env := os.Getenv("NODE_ENV")

	if appErr, ok := err.(*AppError); ok && appErr.IsOperational {
		writeJSON(w, map[string]interface{}{
			"status":  appErr.Status,
			"message": appErr.Message,
			"data": map[string]interface{}{
				"status":  appErr.Status,
				"message": appErr.Message,
			},
		}, appErr.StatusCode)
		return
	}

	if env == "development" {
		writeJSON(w, map[string]interface{}{
			"status":  "fail",
			"message": err.Error(),
			"data":    err,
		}, http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status":  "error",
		"message": err.Error(),
	}, http.StatusInternalServerError)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	err := NewError(404, "Cannot find this url.Please go to Homepage")
	HandleError(w, err)
}

func writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}