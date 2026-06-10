package utils

import (
	"encoding/json"
	"net/http"
)

// ApiResponse 统一响应结构
type ApiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func WriteJSON(w http.ResponseWriter, code int, message string, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(ApiResponse{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
