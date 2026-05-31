package authorization

import (
	"github.com/authzed/authzed-go/v1"
)

type Service struct {
	authzed *authzed.Client
}

type Handler struct {
	svc *Service
}
