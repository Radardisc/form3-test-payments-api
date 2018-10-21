package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/satori/go.uuid"

	"github.com/go-pg/pg/orm"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-pg/pg"
)

var server *http.Server
var db *pg.DB

func TestMain(m *testing.M) {
	db = pg.Connect(&pg.Options{
		User:     "form3",
		Password: "form3",
		Addr:     "database:5432",
	})

	// keep trying until database is ready, as docker container may not be running yet
	for {
		_, err := db.Exec("SELECT 123;")
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}

	if err := db.CreateTable(&Payment{}, &orm.CreateTableOptions{}); err != nil {
		panic(err)
	}

	server = &http.Server{Addr: ":8080", Handler: newAPI(nil)}
	code := m.Run()

	os.Exit(code)
}

func TestGetPaymentsWithEmptyTable(t *testing.T) {

	if _, err := db.Model(&Payment{}).Where("1=1").Delete(); err != nil {
		t.Fatalf("Failed to empty table: %s", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/payments", nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 200 {
		t.Fatalf("Status code was not 200: %d\n", rw.Code)
	}

	if rw.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content type was not application/json")
	}

	var response APIResponse
	err := json.NewDecoder(rw.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode API response: %s", err)
	}

	var payments []Payment
	if err := json.Unmarshal(response.Data, &payments); err != nil {
		t.Fatalf("Failed to decode response to payments slice: %s", err)
	}

	assert.Len(t, payments, 0, "Payments array must be empty when database is empty")
}

func TestGetPaymentsWithOneExistingPayment(t *testing.T) {

	newID := uuid.NewV1()

	// populate table with example payment
	examplePayment := Payment{
		ID: newID,
	}
	db.Insert(&examplePayment)

	req := httptest.NewRequest(http.MethodGet, "/v1/payments", nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 200 {
		t.Fatalf("Status code was not 200: %d\n", rw.Code)
	}

	if rw.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("Content type was not application/json")
	}

	var response APIResponse
	err := json.NewDecoder(rw.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode API response: %s", err)
	}

	var payments []Payment
	if err := json.Unmarshal(response.Data, &payments); err != nil {
		t.Fatalf("Failed to decode response to payments slice: %s", err)
	}

	require.Len(t, payments, 1, "Payments array must contain one payment when database has one payment")
	assert.Equal(t, newID, payments[0].ID)
}
