package ht

import (
	"net/http"
	"time"
)

type LoggingFileServer struct {
	fileServer http.Handler
	log        logger
}

func NewLoggingFileServer(root http.FileSystem) *LoggingFileServer {
	return &LoggingFileServer{
		fileServer: http.FileServer(root),
		log:        &SimpleLogger{},
	}
}

func (s *LoggingFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	timeStart := time.Now()
	sw := NewDetailedResponseWriter(w)
	s.fileServer.ServeHTTP(sw, r)
	s.log.Log(time.Since(timeStart), sw, r)
}
