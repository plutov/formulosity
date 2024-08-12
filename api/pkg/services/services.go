package services

import (
	"fmt"

	"github.com/plutov/formulosity/api/pkg/storage"
)

type Services struct {
	Storage storage.Interface
}

func InitServices() (Services, error) {
	svc := Services{}
	svc.Storage = new(storage.Postgres)

	if err := svc.Storage.Init(); err != nil {
		return svc, fmt.Errorf("unable to init postgres db %w", err)
	}

	return svc, nil
}
