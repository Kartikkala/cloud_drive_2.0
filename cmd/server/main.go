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
	// TODO: Add config for NATS URL change
	nc, err := nats.Connect(nats.DefaultURL)

	if err != nil {
		log.Println("Error in NATS server connection...", err)
		return
	}

	app, err := config.NewApp()
	if err != nil {
		fmt.Println(err.Error())
	}
	authSvc := auth.NewService(app.DB, *app.Cfg)
	// Fill these minio values! Dont forget to turn on minio server!
	minioStorageClient, err := storage.NewMinioStorage("127.0.0.1:9000", *&app.Cfg.Storage.MinioConfig.AccessKeyID, *&app.Cfg.Storage.MinioConfig.SecretAccessKey)
	storageSvc := storage.NewService(app.DB, minioStorageClient)
	artifactsSvcHooks := hooks.NewArtifactsSvcHooks(storageSvc, nc)

	// Register On-Video Hook
	storageSvc.RegisterPutHook(artifactsSvcHooks.OnVideo)

	e := echo.New()
	port := app.Cfg.App.RESTPort

	var jwtMiddlewareFunc echo.MiddlewareFunc = auth.AttachRoutes(e, authSvc)
	storage.AttachRoutes(e, storageSvc, jwtMiddlewareFunc)

	fmt.Println("Starting server on port", port)
	e.Start(fmt.Sprintf("%s:%d", app.Cfg.App.HostAddress, port))
}
