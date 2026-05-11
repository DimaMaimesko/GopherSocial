package main

import (
	"log"

	"github.com/DimaMaimesko/GopherSocial/internal/env"
)

func main() {
	cfg := config{
		addr: env.GetString("ADDR", ":8081"),
	}

	app := &application{
		config: cfg,
	}

	mux := app.mount()

	error := app.run(mux)

	log.Fatal(error)

}
