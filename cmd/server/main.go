package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/sirkartik/cloud_drive_2.0/internal/artifact"
	"github.com/sirkartik/cloud_drive_2.0/internal/auth"
	"github.com/sirkartik/cloud_drive_2.0/internal/config"
	"github.com/sirkartik/cloud_drive_2.0/internal/events"
	"github.com/sirkartik/cloud_drive_2.0/internal/storage"
)

func handleRoot(c echo.Context) error {
	return c.String(http.StatusOK, "This is root!")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := config.NewApp()
	if err != nil {
		fmt.Println(err.Error())
	}
	authSvc := auth.NewService(app.DB, *app.Cfg)
	newJobBroker := events.NewService[*events.Job](2)
	jobProgressBroker := events.NewService[*events.JobProgress](5)
	// Fill these minio values! Dont forget to turn on minio server!
	minioStorageClient, err := storage.NewMinioStorage("127.0.0.1:9000", *&app.Cfg.Storage.MinioConfig.AccessKeyID, *&app.Cfg.Storage.MinioConfig.SecretAccessKey)
	storageSvc := storage.NewService(app.DB, minioStorageClient, newJobBroker)
	artifactSvc := artifact.NewService(app.DB, storageSvc, newJobBroker,jobProgressBroker, 5)

	artifactSvc.StartWorkers(ctx)

	e := echo.New()
	port := app.Cfg.App.RESTPort
	e.GET("/", handleRoot)

	var jwtMiddlewareFunc echo.MiddlewareFunc = auth.AttachRoutes(e, authSvc)
	storage.AttachRoutes(e, storageSvc, jwtMiddlewareFunc)

	fmt.Println("Starting server on port", port)
	e.Start(fmt.Sprintf("%s:%d", app.Cfg.App.HostAddress, port))
}
