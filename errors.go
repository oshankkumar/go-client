package goclient

import (
	"fmt"
	"net/http"
)

// ApiError implements `error` interface. It represents an
// error from a server. It will be used when a server
// responds with 4xx/5xx status code.
type ApiError struct {
	StatusCode int
	RespBody   []byte
	Message    string
}

func (a ApiError) Error() string {
	return fmt.Sprintf("Status: %s, StatusCode: %d ResponseBody: %s Message: %s", http.StatusText(a.StatusCode), a.StatusCode, string(a.RespBody), a.Message)
}
