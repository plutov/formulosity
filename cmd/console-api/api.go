package main

import (
	"os"

	controllers "github.com/plutov/formulosity/pkg/controllers"
	"github.com/plutov/formulosity/pkg/log"
	"github.com/plutov/formulosity/pkg/services"
	"github.com/plutov/formulosity/pkg/surveys"
)

func main() {
	log.Named("console-api")
	logLevel := "info"
	if os.Getenv("LOG_LEVEL") != "" {
		logLevel = os.Getenv("LOG_LEVEL")
	}
	log.SetLogLevel(logLevel)

	svc, err := services.InitServices()
	if err != nil {
		log.WithError(err).Fatal("unable to init dependencies")
	}

	if err := surveys.SyncSurveys(svc); err != nil {
		log.WithError(err).Fatal("unable to sync surveys")
	}

	handler := controllers.NewHandler(svc)
	if err != nil {
		log.WithError(err).Fatal("unable to start server")
	}

	r := controllers.NewRouter(handler)

	if err := r.Start(":8080"); err != nil {
		log.WithError(err).Fatal("shutting down the server")
	}
}
