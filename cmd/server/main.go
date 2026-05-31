package main

import (
	"fmt"
	"log"

	"github.com/authzed/authzed-go/v1"
	"github.com/authzed/grpcutil"
	"github.com/labstack/echo/v4"
	"github.com/nats-io/nats.go"
	"github.com/sirkartik/cloud_drive_2.0/internal/authentication"
	"github.com/sirkartik/cloud_drive_2.0/internal/authorization"
	"github.com/sirkartik/cloud_drive_2.0/internal/config"
	"github.com/sirkartik/cloud_drive_2.0/internal/hooks"
	"github.com/sirkartik/cloud_drive_2.0/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	spicedbEndpoint := fmt.Sprintf("%s:%s", app.Cfg.SpiceDB.URL, app.Cfg.SpiceDB.Port)
	authzedClient, err := authzed.NewClient(
		spicedbEndpoint,
		grpcutil.WithInsecureBearerToken(app.Cfg.SpiceDB.Password),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		log.Println("Error in spiceDB connection...", err)
		return
	}

	authenticationSvc := authentication.NewService(app.DB, *app.Cfg)
	authorizationSvc := authorization.NewService(authzedClient)

	storageSvc := storage.NewService(app.DB, minioStorageClient, *app.Cfg)
	artifactsSvcHooks := hooks.NewArtifactsSvcHooks(storageSvc, nc)

	storageHookLayer := storage.NewHookLayer(storageSvc)

	// Register On-Video Hook
	storageHookLayer.RegisterAfterPutHook(artifactsSvcHooks.OnVideo)

	e := echo.New()
	port := app.Cfg.App.RESTPort

	var jwtMiddlewareFunc echo.MiddlewareFunc = authentication.AttachRoutes(e, authenticationSvc)
	storage.AttachRoutes(e, storageHookLayer, jwtMiddlewareFunc)
	authorization.AttachRoutes(e, authorizationSvc)

	fmt.Println("Starting server on port", port)
	e.Start(fmt.Sprintf("%s:%d", app.Cfg.App.HostAddress, port))
}
