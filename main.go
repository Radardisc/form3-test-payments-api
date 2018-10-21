package main

import (
	"net/http"
	"time"

	"github.com/go-pg/pg"
)

func main() {

	db := pg.Connect(&pg.Options{
		User:     "form3",
		Password: "form3",
		Addr:     "database:5432",
	})

	// keep trying until database is available
	for {
		_, err := db.Exec("SELECT 123;")
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		break
	}

	server := &http.Server{Addr: ":8080", Handler: newAPI(nil)}
	panic(server.ListenAndServe())
}
