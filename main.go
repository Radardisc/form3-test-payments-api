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

type APIResponse struct {
	Data   json.RawMessage `json:"data,omitempty"`
	Links  []Link          `json:"links,omitempty"`
	Errors []string        `json:"errors,omitempty"`
}

type Link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type api struct {
	dataSource *pg.DB
	router     *mux.Router
}

func main() {

	// set up database connection. these should eventually live in a config file but are left here for simplicity.
	db := pg.Connect(&pg.Options{
		User:     "form3",
		Password: "form3",
		Addr:     "database:5432",
	})

	// continually retry until the database connection is successful. once successful, the go-pg package will maintain the connection.
	connectToDatabase(db)

	// provision the database if required.
	if err := provisionDatabase(db); err != nil {
		panic(err)
	}

	// create a new HTTP server in which all requests are handled by the API
	server := &http.Server{Addr: ":8080", Handler: newAPI(nil)}

	// serve continually
	panic(server.ListenAndServe())
}

func connectToDatabase(db *pg.DB) {
	// keep trying until database is available
	for {
		_, err := db.Exec("SELECT 123;") // dummy query to test connection, no ping method on this lib
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}
}

// create the api struct and set up the various handler routes
func newAPI(dataSource *pg.DB) *api {
	api := &api{}

	// create a new mux router and assign handlers to various routes
	api.router = mux.NewRouter()
	api.router.HandleFunc("/v1/payments", api.getPayments).Methods(http.MethodGet)
	api.router.HandleFunc("/v1/payments/{id}", api.getPayment).Methods(http.MethodGet)
	api.router.HandleFunc("/v1/payments", api.createPayment).Methods(http.MethodPost)
	api.router.HandleFunc("/v1/payments/{id}", api.updatePayment).Methods(http.MethodPut)
	api.router.HandleFunc("/v1/payments/{id}", api.deletePayment).Methods(http.MethodDelete)

	// set the db connection on the api
	api.dataSource = dataSource

	return api
}

// implement the http.Handler interface, this just calls the underlying mux handlers serve method
func (api *api) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.router.ServeHTTP(w, r)
}

// business logic for GET /v1/payments endpoint
func (api *api) getPayments(w http.ResponseWriter, r *http.Request) {
	payments := []Payment{}

	// select all payments
	if err := api.dataSource.Model(&payments).Select(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// jsonify the results
	data, err := json.Marshal(payments)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// build the API response (with HATEOAS links)
	response, err := json.Marshal(APIResponse{Data: data, Links: []Link{Link{Rel: "self", Href: "/v1/payments"}}})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// write the response
	w.Header().Add("Content-Type", "application/json")
	w.Write(response)
}

// business logic for GET /v1/payments/{id} endpoint
func (api *api) getPayment(w http.ResponseWriter, r *http.Request) {

	// read the ID from the mux vars
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok { // this should not be possible as muxer will only route requests with an ID
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// parse the supplied UUID
	uuid, err := uuid.FromString(id)
	if err != nil {
		// write an error response indicating the UUID was invalid
		if response, err := json.Marshal(APIResponse{Errors: []string{"Invalid UUID"}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Add("Content-Type", "application/json")
			w.Write(response)
		}
		return
	}

	payment := Payment{
		ID: uuid,
	}

	// select the requested payment from the db
	if err := api.dataSource.Select(&payment); err != nil {
		if err == pg.ErrNoRows {
			if response, err := json.Marshal(APIResponse{Errors: []string{"Payment not found"}}); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusNotFound)
				w.Header().Add("Content-Type", "application/json")
				w.Write(response)
			}
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// encode the query result
	data, err := json.Marshal(payment)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// build the API response (with HATEOAS links)
	response, err := json.Marshal(APIResponse{Data: data, Links: []Link{Link{Rel: "self", Href: fmt.Sprintf("/v1/payments/%s", payment.ID.String())}}})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// write the response
	w.Header().Add("Content-Type", "application/json")
	w.Write(response)
}

// business logic for POST /v1/payments endpoint
func (api *api) createPayment(w http.ResponseWriter, r *http.Request) {

	// read the POSTed payment by decoding it from JSON
	var payment Payment
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payment); err != nil {
		if response, err := json.Marshal(APIResponse{Errors: []string{"Invalid JSON"}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Add("Content-Type", "application/json")
			w.Write(response)
		}
		return
	}

	// select the requested payment from the db
	if err := api.dataSource.Select(&payment); err != pg.ErrNoRows {
		if response, err := json.Marshal(APIResponse{Errors: []string{"Payment already exists with that ID"}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Add("Content-Type", "application/json")
			w.Write(response)
		}
		return
	}

	// insert the record into the db
	if err := api.dataSource.Insert(&payment); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// write response
	w.WriteHeader(http.StatusCreated)
	w.Header().Add("Location", fmt.Sprintf("/v1/payments/%s", payment.ID.String()))
}

// business logic for PUT /v1/payments/{id} endpoint
func (api *api) updatePayment(w http.ResponseWriter, r *http.Request) {

	// grab ID
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok { // the muxer should not assign this handler if the id is missing, so internal error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// parse ID as UUID
	uuid, err := uuid.FromString(id)
	if err != nil {
		if response, err := json.Marshal(APIResponse{Errors: []string{"Invalid UUID"}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(response)
		}
		return
	}

	var payment Payment
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&payment); err != nil {
		if response, err := json.Marshal(APIResponse{Errors: []string{"Invalid JSON"}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(response)
		}
		return
	}

	// ensure the payment being updated matches the one specified in the URL
	if payment.ID.String() != uuid.String() {
		if response, err := json.Marshal(APIResponse{Errors: []string{"Mismatching IDs"}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(response)
		}
		return
	}

	// check the payment exists before editing/replacing it
	existingPayment := Payment{
		ID: uuid,
	}
	if err := api.dataSource.Select(&existingPayment); err != nil {
		if response, err := json.Marshal(APIResponse{Errors: []string{"Payment not found"}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write(response)
		}
		return
	}

	// update the payment in the database
	if err := api.dataSource.Update(&payment); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// write response
	w.WriteHeader(http.StatusCreated)
	w.Header().Add("Location", fmt.Sprintf("/v1/payments/%s", payment.ID.String()))
}

// business logic for DELETE /v1/payments/{id} endpoint
func (api *api) deletePayment(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok { // the muxer should not assign this handler if the id is missing, so internal error
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// parse ID as UUID
	uuid, err := uuid.FromString(id)
	if err != nil {
		if response, err := json.Marshal(APIResponse{Errors: []string{"Invalid UUID"}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write(response)
		}
		return
	}

	// check the payment exists before attempting to delete it
	payment := Payment{
		ID: uuid,
	}
	if err := api.dataSource.Select(&payment); err != nil {
		if response, err := json.Marshal(APIResponse{Errors: []string{"Payment not found"}}); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusNotFound)
			w.Write(response)
		}
		return
	}

	// delete the payment
	if err := api.dataSource.Delete(&payment); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//default to 200 response code
}
