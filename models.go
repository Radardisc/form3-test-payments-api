package main

import uuid "github.com/satori/go.uuid"

type Payment struct {
	Type           string     `json:"type"`
	ID             uuid.UUID  `json:"id" sql:",type:uuid"`
	Version        uint       `json:"version"`
	OrganisationID uuid.UUID  `json:"organisation_id" sql:",type:uuid"`
	Attributes     Attributes `json:"attributes"`
}

type Attributes struct {
	Amount               string             `json:"amount"`
	BeneficiaryParty     BeneficiaryParty   `json:"beneficiary_party"`
	ChargesInformation   ChargesInformation `json:"charges_information"`
	Currency             string             `json:"currency"`
	DebtorParty          DebtorParty        `json:"debtor_party"`
	EndToEndReference    string             `json:"end_to_end_reference"`
	FX                   FX                 `json:"fx"`
	NumericReference     string             `json:"numeric_reference"`
	PaymentID            string             `json:"payment_id"`
	PaymentPurpose       string             `json:"payment_purpose"`
	PaymentScheme        string             `json:"payment_scheme"`
	PaymentType          string             `json:"payment_type"`
	ProcessingDate       string             `json:"processing_date"`
	Reference            string             `json:"reference"`
	SchemePaymentSubType string             `json:"scheme_payment_sub_type"`
	SchemePaymentType    string             `json:"scheme_payment_type"`
	SponsorParty         SponsorParty       `json:"sponsor_party"`
}

type BeneficiaryParty struct {
	*DebtorParty
	AccountType int `json:"account_type"`
}

type DebtorParty struct {
	*SponsorParty
	AccountName       string `json:"account_name"`
	AccountNumberCode string `json:"account_number_code"`
	Address           string `json:"address"`
	Name              string `json:"name"`
}

type SponsorParty struct {
	AccountNumber string `json:"account_number"`
	BankID        string `json:"bank_id"`
	BankIDCode    string `json:"bank_id_code"`
}

type ChargesInformation struct {
	BearerCode              string   `json:"bearer_code"`
	SenderCharges           []Charge `json:"sender_charges"`
	ReceiverChargesAmount   string   `json:"receiver_charges_amount"`
	ReceiverChargesCurrency string   `json:"receiver_charges_currency"`
}

type Charge struct {
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type FX struct {
	ContractReference string `json:"contract_reference"`
	ExchangeRate      string `json:"exchange_rate"`
	OriginalAmount    string `json:"original_amount"`
	OriginalCurrency  string `json:"original_currency"`
}
