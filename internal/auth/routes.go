package auth

import (
	"github.com/labstack/echo/v4"
)

func AttachRoutes(e *echo.Echo, svc *Service){
	handler := NewHandler(svc)
	api := e.Group("/api/auth")
	api.POST("/register", handler.RegisterHandler)
	api.POST("/login", handler.LoginHandler)
}