package auth

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	svc *Service 
}

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) RegisterHandler(c echo.Context) error {
    var req RegisterRequest

    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, "invalid request body")
    }

    err := h.svc.RegisterService(req.Email, req.Username, req.Password)
    if err != nil {
        return c.JSON(http.StatusBadRequest, err.Error())
    }

    return c.JSON(http.StatusCreated, "user registered")
}

func (h *Handler) LoginHandler(c echo.Context) error {
    var req LoginRequest

    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, "invalid request body")
    }

    err, user := h.svc.LoginService(req.Email, req.Password)

    if err != nil {
        return c.JSON(http.StatusUnauthorized, err.Error())
    }

    return c.JSON(http.StatusAccepted, user)
}
