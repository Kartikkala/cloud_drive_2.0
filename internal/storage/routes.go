package storage

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

func AttachRoutes(e *echo.Echo, svc *Service, jwtMiddleware echo.MiddlewareFunc){
	// handler := NewHandler(svc)
	api := e.Group("/api")
	api.Use(jwtMiddleware)
	api.GET("/test", func(c echo.Context) error {
		
		return c.JSON(http.StatusAccepted, map[string] any {
			"res" : c.Get("user"),
		})
	})
	

}