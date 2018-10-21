package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

type api struct {
	dataSource dataSource
	router     *mux.Router
}

func newAPI(dataSource dataSource) *api {
	api := &api{}
	api.router = mux.NewRouter()
	api.dataSource = dataSource
	return api
}

func (api *api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.router.ServeHTTP(w, r)
}
