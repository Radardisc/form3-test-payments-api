package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/go-pg/pg"
)

var server *http.Server

func TestMain(m *testing.M) {
	db := pg.Connect(&pg.Options{
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

	server = &http.Server{Addr: ":8080", Handler: newAPI(nil)}
	code := m.Run()

	os.Exit(code)
}

func TestGetPaymentsWithEmptyTable(t *testing.T) {

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
