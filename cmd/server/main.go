package main

import (
	"fmt"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"
	"github.com/sirkartik/cloud_drive_2.0/internal/auth"
	"github.com/sirkartik/cloud_drive_2.0/internal/config"
	"github.com/sirkartik/cloud_drive_2.0/internal/hooks"
	"github.com/sirkartik/cloud_drive_2.0/internal/storage"
)

func main() {
	app, err := config.NewApp()
	if err != nil {
		fmt.Println(err.Error())
	}

	nc, err := nats.Connect(app.Cfg.NATS.URL)

	if err != nil {
		log.Println("Error in NATS server connection...", err)
		return
	}

	authSvc := auth.NewService(app.DB, *app.Cfg)
	
	minioStorageClient, err := storage.NewMinioStorage(
		app.Cfg.Storage.MinioConfig.Endpoint,
		app.Cfg.Storage.MinioConfig.AccessKeyID,
		app.Cfg.Storage.MinioConfig.SecretAccessKey,
		app.Cfg.Storage.MinioConfig.UseSSL,
	)
	if err != nil {
		log.Println("Error connecting to Minio...", err)
		return
	}

	storageSvc := storage.NewService(app.DB, minioStorageClient, *app.Cfg)
	artifactsSvcHooks := hooks.NewArtifactsSvcHooks(storageSvc, nc)

	storageHookLayer := storage.NewHookLayer(storageSvc)

	// Register On-Video Hook
	storageHookLayer.RegisterAfterPutHook(artifactsSvcHooks.OnVideo)

	e := echo.New()
	port := app.Cfg.App.RESTPort

	var jwtMiddlewareFunc echo.MiddlewareFunc = auth.AttachRoutes(e, authSvc)
	storage.AttachRoutes(e, storageHookLayer, jwtMiddlewareFunc)

	fmt.Println("Starting server on port", port)
	e.Start(fmt.Sprintf("%s:%d", app.Cfg.App.HostAddress, port))
}
