package authorization

import (
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/sirkartik/cloud_drive_2.0/internal/authentication"
)

func NewHandler(svc *Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) CheckPermission(
	permission string,
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Get("user") == nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "User claims missing from request!")
			}

			user, ok := c.Get("user").(*authentication.CustomClaims)

			if !ok {
				log.Println("user context is malformed")
				return echo.NewHTTPError(http.StatusUnauthorized, "Malconfigured user context")
			}

			resourceID := c.Param("id")
			if resourceID == "" {
				return echo.NewHTTPError(http.StatusBadRequest, "Missing node ID in URL")
			}

			allowed, err := h.svc.CheckPermOnResource(
                c.Request().Context(),
                "user", strconv.FormatUint(user.ID, 10),
                "node", resourceID,
                permission, 
                false,
                "",
            )

			if err != nil {
                return echo.NewHTTPError(http.StatusInternalServerError, "Auth system error")
            }
            if !allowed {
                return echo.NewHTTPError(http.StatusForbidden, "Permission denied")
            }

            return next(c)
		}
	}
}
