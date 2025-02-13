package main

import (
	"os"

	controllers "github.com/plutov/formulosity/api/pkg/controllers"
	"github.com/plutov/formulosity/api/pkg/services"
	"github.com/plutov/formulosity/api/pkg/surveys"
)

func main() {
	svc, err := services.InitServices()
	if err != nil {
		svc.Logger.Error("unable to init dependencies", "err", err)
		os.Exit(1)
	}

	if err := surveys.SyncSurveys(svc); err != nil {
		svc.Logger.Error("unable to sync surveys", "err", err)
		os.Exit(1)
	}

	handler := controllers.NewHandler(svc)
	if err != nil {
		svc.Logger.Error("unable to start server", "err", err)
		os.Exit(1)
	}

	r := controllers.NewRouter(handler)

	if err := r.Start(":8080"); err != nil {
		svc.Logger.Info("shutting down the server", "err", err)
	}
}
