package main

import (
	"net/http"

	"github.com/Bendomey/fincore-engine/internal/config"
	"github.com/Bendomey/fincore-engine/internal/db"
	"github.com/Bendomey/fincore-engine/internal/handlers"
	"github.com/Bendomey/fincore-engine/internal/repository"
	"github.com/Bendomey/fincore-engine/internal/router"
	"github.com/Bendomey/fincore-engine/internal/services"
	"github.com/Bendomey/fincore-engine/pkg"
	"github.com/getsentry/raven-go"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := config.Load()

	// init sentry
	if cfg.Env == "production" {
		log.Info("Initializing Sentry...")
		pkg.Sentry(cfg.Sentry.DSN, cfg.Sentry.Environment)
	}

	database, err := db.Connect(cfg)
	if err != nil {
		raven.CaptureError(err, nil)
		log.Fatal("failed to connect db:", err)
	}

	// singleton is efficient.
	validate := validator.New()

	repository := repository.NewRepository(database)
	services := services.NewServices(repository)
	handlers := handlers.NewHandlers(services, validate)

	appCtx := pkg.AppContext{
		DB:         database,
		Config:     cfg,
		Repository: repository,
		Services:   services,
		Handlers:   handlers,
		Validator:  validate,
	}

	r := router.New(appCtx)

	log.Printf("Server running on :%s\n", cfg.Port)

	log.Printf(`[FinCore] :: Server started successfully on http://localhost:%v`, cfg.Port)
	errServer := http.ListenAndServe(":"+cfg.Port, r)

	if errServer != nil {
		raven.CaptureError(errServer, nil)
		log.Fatalf("Error occurred while serving fincore engine, %v", errServer)
	}
}
