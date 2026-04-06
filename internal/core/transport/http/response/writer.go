package http_response

import "net/http"

var (
	statusCodeInitialized = -1
)

type ResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		statusCode:     statusCodeInitialized,
	}
}

func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.ResponseWriter.WriteHeader(statusCode)
	rw.statusCode = statusCode
}

func (rw *ResponseWriter) GetStatusCodeOrPanic() int {
	if rw.statusCode == statusCodeInitialized {
		panic("no status code set")
	}

	return rw.statusCode
}
