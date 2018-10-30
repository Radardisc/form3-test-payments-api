package main

import (
	"reflect"
	"testing"

	"github.com/go-pg/pg"
)

func TestDatabaseProvisioningCreatesTables(t *testing.T) {

	db = pg.Connect(&pg.Options{
		User:     "form3",
		Password: "form3",
		Addr:     "database:5432",
	})

	// provision the database
	provisionDatabase(db)

	models := []interface{}{
		&[]Payment{},
		&[]Attributes{},
		&[]BeneficiaryParty{},
		&[]DebtorParty{},
		&[]SponsorParty{},
		&[]ChargesInformation{},
		&[]Charge{},
		&[]FX{},
	}

	// now check the required tables were created by querying them - this should result in no result and no error
	for _, model := range models {
		if err := db.Model(model).Select(); err != nil {
			t.Errorf("Table was not created for %s", reflect.TypeOf(model).Elem().Elem().Name())
		}
	}

}
