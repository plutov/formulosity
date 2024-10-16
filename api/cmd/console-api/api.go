package main

import (
	"os"

	controllers "github.com/plutov/formulosity/api/pkg/controllers"
	"github.com/plutov/formulosity/api/pkg/log"
	"github.com/plutov/formulosity/api/pkg/middleware"
	"github.com/plutov/formulosity/api/pkg/services"
	"github.com/plutov/formulosity/api/pkg/surveys"
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
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.WithError(err).Fatal("JWT_SECRET environment variable is not set")
	}
	jwtSvc := services.NewJWTService(jwtSecret, svc)

	if err := surveys.SyncSurveys(svc); err != nil {
		log.WithError(err).Fatal("unable to sync surveys")
	}

	handler := controllers.NewHandler(svc, jwtSvc)
	if err != nil {
		log.WithError(err).Fatal("unable to start server")
	}

	r := controllers.NewRouter(handler)
	authMiddleware := middleware.AuthMiddleware(jwtSvc)

	r.Group("/app").Use(authMiddleware)

	if err := r.Start(":8080"); err != nil {
		log.WithError(err).Fatal("shutting down the server")
	}
}
