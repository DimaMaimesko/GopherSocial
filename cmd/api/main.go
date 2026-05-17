package main

import (
	"os"
	"time"

	"github.com/DimaMaimesko/GopherSocial/internal/db"
	"github.com/DimaMaimesko/GopherSocial/internal/env"
	"github.com/DimaMaimesko/GopherSocial/internal/mailer"
	"github.com/DimaMaimesko/GopherSocial/internal/store"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/joho/godotenv"
)

const version = "0.0.1"

// @title			GopherSocial API
// @version		1.0
// @description	API documentation for GopherSocial.
// @host			localhost:8082
// @BasePath		/v1
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
			exp:       time.Hour * 24 * 3, // 3 days
			fromEmail: env.GetString("FROM_EMAIL", ""),
			sendGrid: sendGridConfig{
				apiKey: env.GetString("SENDGRID_API_KEY", ""),
			},
			mailTrap: mailTrapConfig{
				apiKey:    env.GetString("MAILTRAP_API_KEY", "24fe2dadf5e4a531120fa99d55350a84"),
				sandboxID: env.GetString("MAILTRAP_SANDBOX_ID", "4634866"),
			},
		},
		frontendURL: env.GetString("FRONTEND_URL", "http://localhost:4000"),
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

	// Mailer
	// mailer := mailer.NewSendgrid(cfg.mail.sendGrid.apiKey, cfg.mail.fromEmail)
	mailtrap, err := mailer.NewMailTrapClient(
		cfg.mail.mailTrap.apiKey,
		cfg.mail.mailTrap.sandboxID,
		cfg.mail.fromEmail,
	)
	if err != nil {
		logger.Fatal(err)
	}

	app := &application{
		config: cfg,
		store:  st,
		logger: logger,
		//mailer: mailer,
		mailer: mailtrap,
	}

	mux := app.mount()

	logger.Fatal(app.run(mux))

}
