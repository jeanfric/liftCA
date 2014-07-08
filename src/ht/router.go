package ht

import (
	"net/http"

	"github.com/gorilla/mux"
)

type myRouter struct {
	router *mux.Router
}

func NewRouter() *myRouter {
	return &myRouter{
		router: mux.NewRouter(),
	}
}

func (m *myRouter) Handle(method, path string, h http.Handler) {
	m.router.Handle(path, h).Methods(method)
}

func (m *myRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.router.ServeHTTP(w, r)
}
