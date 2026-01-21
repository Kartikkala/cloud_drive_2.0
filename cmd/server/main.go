package main

import (
	"fmt"
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/sirkartik/cloud_drive_2.0/internal/config"
	"github.com/sirkartik/cloud_drive_2.0/internal/auth"
)


func handleRoot(c echo.Context) error {
	return c.String(http.StatusOK, "This is root!");
}

func main() {
	app, err := config.NewApp()
	if err != nil {
		fmt.Println(err.Error())
	}
	authSvc := auth.NewService(app.DB)
	
	e := echo.New()
	port := app.Cfg.App.RESTPort
	e.GET("/", handleRoot)
	auth.AttachRoutes(e, authSvc)
	fmt.Println("Starting server on port", port)
	e.Start(fmt.Sprintf("%s:%d", app.Cfg.App.HostAddress, port))
}
