package utils

import (
	"encoding/json"
	"net/http"
)

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func Success(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

func Error(w http.ResponseWriter, message string, status int) {
	JSON(w, status, map[string]interface{}{
		"status":  "fail",
		"message": message,
	})
}

func ServerError(w http.ResponseWriter, message string) {
	JSON(w, http.StatusInternalServerError, map[string]interface{}{
		"status":  "error",
		"message": message,
	})
}