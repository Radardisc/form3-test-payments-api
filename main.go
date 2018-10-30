package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
)

func main() {

	db := pg.Connect(&pg.Options{
		User:     "form3",
		Password: "form3",
		Addr:     "database:5432",
	})

	connectToDatabase(db)

	if err := provisionDatabase(db); err != nil {
		panic(err)
	}

	server := &http.Server{Addr: ":8080", Handler: newAPI(nil)}
	panic(server.ListenAndServe())
}

func connectToDatabase(db *pg.DB) {
	// keep trying until database is available
	for {
		_, err := db.Exec("SELECT 123;")
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}
}

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
	api.router.HandleFunc("/v1/payments", api.createPayment).Methods(http.MethodPost)
	api.router.HandleFunc("/v1/payments/{id}", api.updatePayment).Methods(http.MethodPut)
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

func (api *api) createPayment(w http.ResponseWriter, r *http.Request) {

	var payment Payment
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payment); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := api.dataSource.Insert(&payment); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Add("Location", fmt.Sprintf("/v1/payments/%s", payment.ID.String()))
}

func (api *api) updatePayment(w http.ResponseWriter, r *http.Request) {

	var payment Payment
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payment); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := api.dataSource.Update(&payment); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Add("Location", fmt.Sprintf("/v1/payments/%s", payment.ID.String()))
}
