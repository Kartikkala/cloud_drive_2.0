package storage

import (
	// "github.com/labstack/echo/v4"
)

type Handler struct {
	svc Service 
}

func NewHandler(svc Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

