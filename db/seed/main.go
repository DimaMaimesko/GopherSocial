package main

import (
	"log"

	"github.com/DimaMaimesko/GopherSocial/internal/db"
	"github.com/DimaMaimesko/GopherSocial/internal/env"
	"github.com/DimaMaimesko/GopherSocial/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/socialnetwork?sslmode=disable")
	conn, err := db.New(addr, 3, 3, "15m")
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	store := store.NewPostgresStorage(conn)

	db.Seed(store, conn)
}
