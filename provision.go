package main

import (
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
)

func provisionDatabase(db *pg.DB) error {

	if err := db.CreateTable(&Payment{}, &orm.CreateTableOptions{}); err != nil {
		return err
	}

	if err := db.CreateTable(&Attributes{}, &orm.CreateTableOptions{}); err != nil {
		return err
	}

	if err := db.CreateTable(&BeneficiaryParty{}, &orm.CreateTableOptions{}); err != nil {
		return err
	}

	if err := db.CreateTable(&DebtorParty{}, &orm.CreateTableOptions{}); err != nil {
		return err
	}

	if err := db.CreateTable(&SponsorParty{}, &orm.CreateTableOptions{}); err != nil {
		return err
	}

	if err := db.CreateTable(&ChargesInformation{}, &orm.CreateTableOptions{}); err != nil {
		return err
	}

	if err := db.CreateTable(&Charge{}, &orm.CreateTableOptions{}); err != nil {
		return err
	}

	if err := db.CreateTable(&FX{}, &orm.CreateTableOptions{}); err != nil {
		return err
	}

	return nil
}
