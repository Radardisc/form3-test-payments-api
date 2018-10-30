package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/satori/go.uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/go-pg/pg"
)

var server *http.Server
var db *pg.DB

func TestMain(m *testing.M) {

	// connect to the database, create the required tables and run the API server to test against

	db = pg.Connect(&pg.Options{
		User:     "form3",
		Password: "form3",
		Addr:     "database:5432",
	})

	connectToDatabase(db)

	provisionDatabase(db)

	server = &http.Server{Addr: ":8080", Handler: newAPI(db)}
	code := m.Run()

	os.Exit(code)
}

func emptyDatabase(t *testing.T) {

	// remove all rows from all of the tables

	models := []interface{}{
		&Payment{},
		&Attributes{},
		&BeneficiaryParty{},
		&DebtorParty{},
		&SponsorParty{},
		&ChargesInformation{},
		&Charge{},
		&FX{},
	}

	for _, model := range models {
		if _, err := db.Model(model).Where("1=1").Delete(); err != nil {
			t.Fatalf("Failed to empty table: %s", err)
		}
	}
}

func createExamplePayment() Payment {
	return Payment{
		Type:           "Payment",
		ID:             uuid.NewV1(),
		Version:        0,
		OrganisationID: uuid.NewV1(),
		Attributes: Attributes{
			Amount: "100.00",
			BeneficiaryParty: BeneficiaryParty{
				DebtorParty: &DebtorParty{
					SponsorParty: &SponsorParty{
						AccountNumber: "12345678",
						BankID:        "203301",
						BankIDCode:    "GBDSC",
					},
					AccountName:       "L Galvin",
					AccountNumberCode: "IBAN",
					Name:              "Liam Galvin",
					Address:           "123 Main Street",
				},
				AccountType: 0,
			},
			ChargesInformation: ChargesInformation{
				BearerCode: "SHAR",
				SenderCharges: []Charge{
					{
						Amount:   "0.50",
						Currency: "GBP",
					},
					{
						Amount:   "0.10",
						Currency: "USD",
					},
				},
				ReceiverChargesAmount:   "1.00",
				ReceiverChargesCurrency: "GBP",
			},
			Currency: "GBP",
			DebtorParty: DebtorParty{
				SponsorParty: &SponsorParty{
					AccountNumber: "77777777",
					BankID:        "203301",
					BankIDCode:    "GBDSC",
				},
				AccountName:       "Mangoes Inc",
				AccountNumberCode: "IBAN",
				Name:              "Mangoes Incorporated",
				Address:           "124 Main Street",
			},
			EndToEndReference:    "payment for mangoes",
			NumericReference:     "1012321",
			PaymentID:            "123456789012345678",
			PaymentPurpose:       "Paying for goods/services",
			PaymentScheme:        "FPS",
			PaymentType:          "Credit",
			ProcessingDate:       "2017-01-18",
			Reference:            "Payment for Em's mangoes",
			SchemePaymentSubType: "InternetBanking",
			SchemePaymentType:    "ImmediatePayment",
			SponsorParty: SponsorParty{
				AccountNumber: "10101010",
				BankID:        "203302",
				BankIDCode:    "GBDSC",
			},
		},
	}
}

func TestGetPaymentsWithEmptyTable(t *testing.T) {

	emptyDatabase(t)

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

	assert.EqualValues(t, []Link{Link{Rel: "self", Href: "/v1/payments"}}, response.Links)

	assert.Len(t, payments, 0, "Payments array must be empty when database is empty")
}

func TestGetPaymentsWithOneExistingPayment(t *testing.T) {

	emptyDatabase(t)

	// populate table with example payment
	examplePayment := createExamplePayment()
	if err := db.Insert(&examplePayment); err != nil {
		t.Fatal(err)
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

	require.Len(t, payments, 1, "Payments array must contain one payment when database has one payment")
	assert.EqualValues(t, examplePayment, payments[0])

	assert.EqualValues(t, []Link{Link{Rel: "self", Href: "/v1/payments"}}, response.Links)
}

func TestGetPaymentsWithMultipleExistingPayments(t *testing.T) {

	emptyDatabase(t)

	// populate table with example payment
	examplePayments := []Payment{
		createExamplePayment(),
		{
			Type:           "Payment",
			ID:             uuid.NewV1(),
			Version:        0,
			OrganisationID: uuid.NewV1(),
			Attributes: Attributes{
				Amount: "100.00",
				BeneficiaryParty: BeneficiaryParty{
					DebtorParty: &DebtorParty{
						SponsorParty: &SponsorParty{
							AccountNumber: "12345678",
							BankID:        "203301",
							BankIDCode:    "GBDSC",
						},
						AccountName:       "L Galvin",
						AccountNumberCode: "IBAN",
						Name:              "Liam Galvin",
						Address:           "123 Main Street",
					},
					AccountType: 0,
				},
				ChargesInformation: ChargesInformation{
					BearerCode: "SHAR",
					SenderCharges: []Charge{
						{
							Amount:   "0.50",
							Currency: "GBP",
						},
						{
							Amount:   "0.10",
							Currency: "USD",
						},
					},
					ReceiverChargesAmount:   "1.00",
					ReceiverChargesCurrency: "GBP",
				},
				Currency: "GBP",
				DebtorParty: DebtorParty{
					SponsorParty: &SponsorParty{
						AccountNumber: "77777777",
						BankID:        "203301",
						BankIDCode:    "GBDSC",
					},
					AccountName:       "Mangoes Inc",
					AccountNumberCode: "IBAN",
					Name:              "Mangoes Incorporated",
					Address:           "124 Main Street",
				},
				EndToEndReference:    "payment for mangoes",
				NumericReference:     "1012321",
				PaymentID:            "123456789012345678",
				PaymentPurpose:       "Paying for goods/services",
				PaymentScheme:        "FPS",
				PaymentType:          "Credit",
				ProcessingDate:       "2017-01-18",
				Reference:            "Payment for Em's mangoes",
				SchemePaymentSubType: "InternetBanking",
				SchemePaymentType:    "ImmediatePayment",
				SponsorParty: SponsorParty{
					AccountNumber: "10101010",
					BankID:        "203302",
					BankIDCode:    "GBDSC",
				},
			},
		},
	}
	if err := db.Insert(&examplePayments); err != nil {
		t.Fatal(err)
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

	require.Len(t, payments, 2, "Payments array must contain two payments when database has two payments")
	assert.EqualValues(t, examplePayments, payments)

	assert.EqualValues(t, []Link{Link{Rel: "self", Href: "/v1/payments"}}, response.Links)
}

func TestGetSinglePaymentWithOneExistingPayment(t *testing.T) {

	emptyDatabase(t)

	// populate table with example payment
	examplePayment := createExamplePayment()
	if err := db.Insert(&examplePayment); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/payments/%s", examplePayment.ID.String()), nil)
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

	var payment Payment
	if err := json.Unmarshal(response.Data, &payment); err != nil {
		t.Fatalf("Failed to decode response to payment: %s", err)
	}

	assert.EqualValues(t, examplePayment, payment)

	assert.EqualValues(t, []Link{Link{Rel: "self", Href: fmt.Sprintf("/v1/payments/%s", examplePayment.ID.String())}}, response.Links)
}

func TestGetSinglePaymentForNonExistingPayment(t *testing.T) {

	emptyDatabase(t)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/payments/%s", uuid.NewV1().String()), nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 404 {
		t.Fatalf("Status code was not 404: %d\n", rw.Code)
	}
	var response APIResponse
	err := json.NewDecoder(rw.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode API response: %s", err)
	}
	assert.EqualValues(t, []string{"Payment not found"}, response.Errors)
}

func TestGetSinglePaymentForInvalidUUID(t *testing.T) {

	emptyDatabase(t)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/payments/%s", "blah"), nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 400 {
		t.Fatalf("Status code was not 400: %d\n", rw.Code)
	}
	var response APIResponse
	err := json.NewDecoder(rw.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode API response: %s", err)
	}
	assert.EqualValues(t, []string{"Invalid UUID"}, response.Errors)
}

func TestGetSinglePaymentForNonExistingPaymentWhenOtherPaymentExists(t *testing.T) {

	emptyDatabase(t)

	// populate table with example payment
	examplePayment := createExamplePayment()
	if err := db.Insert(&examplePayment); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/payments/%s", uuid.NewV1().String()), nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 404 {
		t.Fatalf("Status code was not 404: %d\n", rw.Code)
	}
	var response APIResponse
	err := json.NewDecoder(rw.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode API response: %s", err)
	}
	assert.EqualValues(t, []string{"Payment not found"}, response.Errors)
}

func TestCreateSinglePayment(t *testing.T) {

	emptyDatabase(t)

	// populate table with example payment
	examplePayment := createExamplePayment()

	jsonBytes, err := json.Marshal(examplePayment)
	require.Nil(t, err)

	req := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 201 {
		t.Fatalf("Status code was not 201: %d\n", rw.Code)
	}
	assert.Equal(t, fmt.Sprintf("/v1/payments/%s", examplePayment.ID.String()), rw.Header().Get("Location"))

	actualPayment := Payment{
		ID: examplePayment.ID,
	}
	err = db.Select(&actualPayment)
	require.Nil(t, err)

	assert.EqualValues(t, examplePayment, actualPayment)
}

func TestCreateSinglePaymentWithInvalidJSON(t *testing.T) {

	emptyDatabase(t)

	jsonBytes := []byte("{ bad json }")
	req := httptest.NewRequest(http.MethodPost, "/v1/payments", bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 400 {
		t.Fatalf("Status code was not 400: %d\n", rw.Code)
	}
	var response APIResponse
	err := json.NewDecoder(rw.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode API response: %s", err)
	}
	assert.EqualValues(t, []string{"Invalid JSON"}, response.Errors)
}

func TestUpdatePayment(t *testing.T) {

	emptyDatabase(t)

	// populate table with example payment
	examplePayment := createExamplePayment()
	if err := db.Insert(&examplePayment); err != nil {
		t.Fatal(err)
	}

	examplePayment.Attributes.BeneficiaryParty.Name = "John Smith"

	jsonBytes, err := json.Marshal(examplePayment)
	require.Nil(t, err)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/payments/%s", examplePayment.ID), bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 201 {
		t.Fatalf("Status code was not 201: %d\n", rw.Code)
	}
	assert.Equal(t, fmt.Sprintf("/v1/payments/%s", examplePayment.ID.String()), rw.Header().Get("Location"))

	actualPayment := Payment{
		ID: examplePayment.ID,
	}
	err = db.Select(&actualPayment)
	require.Nil(t, err)

	assert.EqualValues(t, examplePayment, actualPayment)
}

func TestUpdateSinglePaymentWithIDThatDoesNotMatchURL(t *testing.T) {

	emptyDatabase(t)

	// populate table with example payment
	examplePayment := createExamplePayment()
	if err := db.Insert(&examplePayment); err != nil {
		t.Fatal(err)
	}

	jsonBytes, err := json.Marshal(examplePayment)
	require.Nil(t, err)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/payments/%s", uuid.NewV1().String()), bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 400 {
		t.Fatalf("Status code was not 400: %d\n", rw.Code)
	}
	var response APIResponse
	if err := json.NewDecoder(rw.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode API response: %s", err)
	}
	assert.EqualValues(t, []string{"Mismatching IDs"}, response.Errors)
}

func TestUpdateNonExistentPayment(t *testing.T) {

	emptyDatabase(t)

	// populate table with example payment
	examplePayment := createExamplePayment()

	jsonBytes, err := json.Marshal(examplePayment)
	require.Nil(t, err)

	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/payments/%s", examplePayment.ID.String()), bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 404 {
		t.Fatalf("Status code was not 404: %d\n", rw.Code)
	}
	var response APIResponse
	if err := json.NewDecoder(rw.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode API response: %s", err)
	}
	assert.EqualValues(t, []string{"Payment not found"}, response.Errors)
}

func TestUpdateSinglePaymentWithInvalidJSON(t *testing.T) {

	emptyDatabase(t)

	// populate table with example payment
	examplePayment := createExamplePayment()
	if err := db.Insert(&examplePayment); err != nil {
		t.Fatal(err)
	}

	jsonBytes := []byte("{ bad json }")
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/v1/payments/%s", examplePayment.ID.String()), bytes.NewBuffer(jsonBytes))
	req.Header.Set("Content-Type", "application/json")
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 400 {
		t.Fatalf("Status code was not 400: %d\n", rw.Code)
	}
	var response APIResponse
	err := json.NewDecoder(rw.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode API response: %s", err)
	}
	assert.EqualValues(t, []string{"Invalid JSON"}, response.Errors)
}

func TestDeletePayment(t *testing.T) {

	emptyDatabase(t)

	// populate table with example payment
	examplePayment := createExamplePayment()
	if err := db.Insert(&examplePayment); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/payments/%s", examplePayment.ID), nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 200 {
		t.Fatalf("Status code was not 200: %d\n", rw.Code)
	}

	actualPayment := Payment{
		ID: examplePayment.ID,
	}
	err := db.Select(&actualPayment)
	assert.Equal(t, pg.ErrNoRows, err)
}

func TestDeleteNonExistingPayment(t *testing.T) {

	emptyDatabase(t)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/v1/payments/%s", uuid.NewV1().String()), nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 404 {
		t.Fatalf("Status code was not 404: %d\n", rw.Code)
	}
	var response APIResponse
	err := json.NewDecoder(rw.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Failed to decode API response: %s", err)
	}
	assert.EqualValues(t, []string{"Payment not found"}, response.Errors)

}
