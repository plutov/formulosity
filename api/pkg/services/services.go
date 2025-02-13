package services

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/plutov/formulosity/api/pkg/storage"
)

type Services struct {
	Storage     storage.Interface
	FileStorage storage.FileInterface
	Logger      *slog.Logger
}

func InitServices() (Services, error) {
	svc := Services{
		Logger: slog.New(slog.NewJSONHandler(os.Stdout, nil)),
	}
	switch os.Getenv("DATABASE_TYPE") {
	case "postgres":
		svc.Storage = new(storage.Postgres)
	case "sqlite":
		svc.Storage = new(storage.Sqlite)
	default:
		return svc, fmt.Errorf("unknown database type")
	}

	if err := svc.Storage.Init(); err != nil {
		return svc, fmt.Errorf("unable to init db %w", err)
	}

	svc.FileStorage = new(storage.File)
	if err := svc.FileStorage.Init(); err != nil {
		return svc, fmt.Errorf("unable to init file storage %w", err)
	}

	return svc, nil
}
