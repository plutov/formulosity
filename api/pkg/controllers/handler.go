package controllers

import (
	"github.com/plutov/formulosity/api/pkg/services"
)

type Handler struct {
	services.Services
	JWTService services.JwtService
}

func NewHandler(svc services.Services, jwtSvc services.JwtService) *Handler {
	h := &Handler{
		Services:   svc,
		JWTService: jwtSvc,
	}
	return h
}
