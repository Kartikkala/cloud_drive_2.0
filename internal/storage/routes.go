package storage

import (
	"github.com/labstack/echo/v4"
)

func AttachRoutes(e *echo.Echo, svc *Service, jwtMiddleware echo.MiddlewareFunc){
	handler := NewHandler(*svc)
	api := e.Group("/api")
	api.Use(jwtMiddleware)
	api.POST("/upload", handler.Upload)
	api.POST("/download", handler.Download)
	api.POST("/list", handler.List)
	api.POST("/mkdir", handler.CreateDirectoryNode)
	api.POST("/copy", handler.Copy)
	api.POST("/move", handler.Move)
	api.POST("/delete", handler.Delete)
}