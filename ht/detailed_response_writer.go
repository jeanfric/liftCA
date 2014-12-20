package ht

import (
	"net/http"
)

// statusResponseWriter implements the http.ResponseWriter by wrapping another implementation,
// but in addition provides access to the status code and size of the response.
type DetailedResponseWriter struct {
	responseWriter http.ResponseWriter
	status         int
	size           int
}

func NewDetailedResponseWriter(w http.ResponseWriter) *DetailedResponseWriter {
	return &DetailedResponseWriter{
		responseWriter: w,
		status:         http.StatusOK,
		size:           0,
	}
}

func (w *DetailedResponseWriter) Header() http.Header {
	return w.responseWriter.Header()
}

func (w *DetailedResponseWriter) Write(bytes []byte) (size int, err error) {
	size, err = w.responseWriter.Write(bytes)
	w.size += size
	return
}

func (w *DetailedResponseWriter) WriteHeader(statusCode int) {
	w.responseWriter.WriteHeader(statusCode)
	w.status = statusCode
}

func (w *DetailedResponseWriter) Size() int {
	return w.size
}

func (w *DetailedResponseWriter) Status() int {
	return w.status
}
