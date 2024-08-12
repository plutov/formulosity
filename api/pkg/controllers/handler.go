package controllers

import (
	"github.com/plutov/formulosity/api/pkg/services"
)

type Handler struct {
	services.Services
}

func NewHandler(svc services.Services) *Handler {
	h := &Handler{
		Services: svc,
	}
	return h
}
