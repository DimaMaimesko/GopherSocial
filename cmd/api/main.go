package main

import (
	"os"
	"time"

	"github.com/DimaMaimesko/GopherSocial/internal/db"
	"github.com/DimaMaimesko/GopherSocial/internal/env"
	"github.com/DimaMaimesko/GopherSocial/internal/store"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/joho/godotenv"
)

const version = "0.0.1"

func main() {

	if err := os.MkdirAll("logs", 0755); err != nil {
		panic(err)
	}

	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Encoding = "console"
	loggerConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	loggerConfig.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	loggerConfig.OutputPaths = []string{"stdout", "logs/app.log"}
	loggerConfig.ErrorOutputPaths = []string{"stderr", "logs/app-error.log"}

	logger := zap.Must(loggerConfig.Build()).Sugar()
	defer logger.Sync()

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
		mail: mailConfig{
			exp: time.Hour * 24 * 3, // 3 days
		},
	}

	db, err := db.New(
		cfg.db.addr,
		cfg.db.maxOpenConns,
		cfg.db.maxIdleConns,
		cfg.db.maxIdleTime,
	)
	if err != nil {
		logger.Fatal(err)
	}

	defer db.Close()
	logger.Info("database connection pool established")
	st := store.NewPostgresStorage(db)

	app := &application{
		config: cfg,
		store:  st,
		logger: logger,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))

}
