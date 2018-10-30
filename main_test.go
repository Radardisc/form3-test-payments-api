package main

import (
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

	assert.Len(t, payments, 0, "Payments array must be empty when database is empty")
}

func TestGetPaymentsWithOneExistingPayment(t *testing.T) {

	emptyDatabase(t)

	newID := uuid.NewV1()
	organisationID := uuid.NewV1()

	// populate table with example payment
	examplePayment := Payment{
		Type:           "Payment",
		ID:             newID,
		Version:        0,
		OrganisationID: organisationID,
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
}

func TestGetPaymentsWithMultipleExistingPayments(t *testing.T) {

	emptyDatabase(t)

	// populate table with example payment
	examplePayments := []Payment{
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
}

func TestGetSinglePaymentWithOneExistingPayment(t *testing.T) {

	emptyDatabase(t)

	// populate table with example payment
	examplePayment := Payment{
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
}

func TestGetSinglePaymentForNonExistingPayment(t *testing.T) {

	emptyDatabase(t)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/payments/%s", uuid.NewV1().String()), nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 404 {
		t.Fatalf("Status code was not 404: %d\n", rw.Code)
	}
}

func TestGetSinglePaymentForInvalidUUID(t *testing.T) {

	emptyDatabase(t)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/payments/%s", "blah"), nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 400 {
		t.Fatalf("Status code was not 400: %d\n", rw.Code)
	}
}

func TestGetSinglePaymentForNonExistingPaymentWhenOtherPaymentExists(t *testing.T) {

	emptyDatabase(t)

	// populate table with example payment
	examplePayment := Payment{
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
	if err := db.Insert(&examplePayment); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/v1/payments/%s", uuid.NewV1().String()), nil)
	rw := httptest.NewRecorder()
	server.Handler.ServeHTTP(rw, req)
	if rw.Code != 404 {
		t.Fatalf("Status code was not 404: %d\n", rw.Code)
	}
}
