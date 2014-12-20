package ht

import (
	"log"
	"net/http"
	"time"
)

type logger interface {
	Log(time.Duration, *DetailedResponseWriter, *http.Request)
}

type SimpleLogger struct{}

func (l *SimpleLogger) Log(t time.Duration, w *DetailedResponseWriter, r *http.Request) {
	log.Printf("%v: %v %v => %v (%v, %vB)", r.RemoteAddr, r.Method, r.URL, w.Status(), t, w.Size())
}
