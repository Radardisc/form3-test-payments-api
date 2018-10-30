package main

import (
	"encoding/json"
	"net/http"

	"github.com/satori/go.uuid"

	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
)

type APIResponse struct {
	Data json.RawMessage `json:"data"`
}

type api struct {
	dataSource *pg.DB
	router     *mux.Router
}

func newAPI(dataSource *pg.DB) *api {
	api := &api{}
	api.router = mux.NewRouter()
	api.router.HandleFunc("/v1/payments", api.getPayments).Methods(http.MethodGet)
	api.router.HandleFunc("/v1/payments/{id}", api.getPayment).Methods(http.MethodGet)
	api.dataSource = dataSource
	return api
}

func (api *api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.router.ServeHTTP(w, r)
}

func (api *api) getPayments(w http.ResponseWriter, r *http.Request) {
	payments := []Payment{}

	if err := api.dataSource.Model(&payments).Select(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(payments)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(APIResponse{Data: data})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(response)
}

func (api *api) getPayment(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	uuid, err := uuid.FromString(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	payment := Payment{
		ID: uuid,
	}

	if err := api.dataSource.Select(&payment); err != nil {
		if err == pg.ErrNoRows {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(payment)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	response, err := json.Marshal(APIResponse{Data: data})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(response)
}
