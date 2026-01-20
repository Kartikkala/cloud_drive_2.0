package main

import (
	"fmt"
	"net/http"
	"github.com/labstack/echo/v4"
	"github.com/sirkartik/cloud_drive_2.0/internal/config"
	"github.com/sirkartik/cloud_drive_2.0/internal/server"
)

func handleRoot(c echo.Context) error {
	return c.String(http.StatusOK, "This is root!");
}

func main() {
	app, err := config.NewApp()
	if err != nil {
		fmt.Println(err.Error())
	}

	server := server.Server{
		App: app,
	}
	
	e := echo.New()
	port := app.Cfg.App.RESTPort
	e.GET("/", handleRoot)
	fmt.Println("Starting server on port", port)
	e.Start(fmt.Sprintf("%s:%d", app.Cfg.App.HostAddress, port))
}
