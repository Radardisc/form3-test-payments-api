package main

import uuid "github.com/satori/go.uuid"

type Payment struct {
	Type           string     `json:"type"`
	ID             uuid.UUID  `json:"id"`
	Version        uint       `json:"version"`
	OrganisationID uuid.UUID  `json:"organisation_id"`
	Attributes     Attributes `json:"attributes"`
}

type Attributes struct {
}
