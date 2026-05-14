package main

import (
	"log"

	"github.com/DimaMaimesko/GopherSocial/internal/db"
	"github.com/DimaMaimesko/GopherSocial/internal/env"
	"github.com/DimaMaimesko/GopherSocial/internal/store"

	"github.com/joho/godotenv"
)

const version = "0.0.1"

func main() {

	_ = godotenv.Load(".envrc")
	//log.Println("DB_ADDR =", os.Getenv("DB_ADDR"))

	cfg := config{
		addr: env.GetString("ADDR", ":8080"),
		db: dbConfig{
			addr:         env.GetString("DB_ADDR", "postgres://admin:adminpassword@localhost/socialnetwork?sslmode=disable"),
			maxOpenConns: env.GetInt("DB_MAX_OPEN_CONNS", 30),
			maxIdleConns: env.GetInt("DB_MAX_IDLE_CONNS", 30),
			maxIdleTime:  env.GetString("DB_MAX_IDLE_TIME", "15m"),
		},
		env: env.GetString("ENV", "development"),
	}

	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		log.Panic(err)
	}

	defer db.Close()
	log.Println("database connection pool established")
	st := store.NewPostgresStorage(db)

	app := &application{
		config: cfg,
		store:  st,
	}

	mux := app.mount()

	error := app.run(mux)

	log.Fatal(error)

}
