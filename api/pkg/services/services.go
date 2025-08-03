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
		Storage: new(storage.Postgres),
	}

	if err := svc.Storage.Init(); err != nil {
		return svc, fmt.Errorf("unable to init db %w", err)
	}

	svc.FileStorage = &storage.File{
		Logger: svc.Logger,
	}
	if err := svc.FileStorage.Init(); err != nil {
		return svc, fmt.Errorf("unable to init file storage %w", err)
	}

	return svc, nil
}
