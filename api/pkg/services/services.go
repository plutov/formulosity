package services

import (
	"fmt"
	"os"

	"github.com/plutov/formulosity/api/pkg/storage"
)

type Services struct {
	Storage     storage.Interface
	FileStorage storage.FileInterface
}

func InitServices() (Services, error) {
	svc := Services{}
	switch os.Getenv("DATABASE_TYPE") {
	case "postgres":
		svc.Storage = new(storage.Postgres)
	case "sqlite":
		svc.Storage = new(storage.Sqlite)
	case "mysql":
		svc.Storage = new(storage.Mysql)
	default:
		return svc, fmt.Errorf("unknown database type: %s", os.Getenv("DATABASE_TYPE"))
	}

	if err := svc.Storage.Init(); err != nil {
		return svc, fmt.Errorf("unable to init db - %w", err)
	}

	svc.FileStorage = new(storage.File)
	if err := svc.FileStorage.Init(); err != nil {
		return svc, fmt.Errorf("unable to init file storage %w", err)
	}

	return svc, nil
}
