package storage

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/sirkartik/cloud_drive_2.0/internal/auth"
)

type Handler struct {
	svc Service
}

func NewHandler(svc Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h Handler) Download(c echo.Context) error {
	var req DLoad
	ctx := c.Request().Context()
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "missing id param in request body")
	}
	id, err := uuid.Parse(req.NodeID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, "invalid id param")
	}
	var user *auth.CustomClaims = c.Get("user").(*auth.CustomClaims)

	stream, node, err := h.svc.GetData(ctx, id, user.ID)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, "error getting data from storage service")
	}

	c.Response().Header().Set(
		echo.HeaderContentDisposition,
		fmt.Sprintf(`attachment; filename="%s"`, node.Name),
	)
	c.Response().Header().Set(echo.HeaderContentType, "application/octet-stream")
	c.Response().WriteHeader(http.StatusOK)
	_, err = io.Copy(c.Response().Writer, stream)
	return err
}

func (h Handler) Upload(c echo.Context) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusBadRequest, "missing file")
	}
	ctx := c.Request().Context()

	file, err := fileHeader.Open()

	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, "cannot open file")
	}

	defer file.Close()
	var user *auth.CustomClaims = c.Get("user").(*auth.CustomClaims)
	var parentNodeId string = c.Request().Header.Get("parent_id")

	filename := fileHeader.Filename
	size := fileHeader.Size
	parentId, err := uuid.Parse(parentNodeId)
	if err != nil {
		parentId = uuid.Nil
	}
	err = h.svc.Put(ctx, user.ID, parentId, filename, uint64(size), file)

	if err != nil {
		log.Println(err)
		return c.JSON(http.StatusInternalServerError, "error writing file to the storage")
	}
	return c.NoContent(http.StatusCreated)
}

func (h Handler) List(c echo.Context) error {
	var req ListNodes
	ctx := c.Request().Context()

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "invalid request body")
	}
	var user *auth.CustomClaims = c.Get("user").(*auth.CustomClaims)

	parentId, _ := uuid.Parse(req.ParentID)
	nodeList, err := h.svc.ListNodes(ctx, parentId, user.ID)

	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusInternalServerError, "error fetching node list")
	}

	return c.JSON(http.StatusAccepted, map[string]interface{}{
		"list": &nodeList,
	})
}

func (h Handler) CreateDirectoryNode(c echo.Context) error {
	var req Mkdir
	ctx := c.Request().Context()

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "invalid request body")
	}
	var user *auth.CustomClaims = c.Get("user").(*auth.CustomClaims)
	parentId, _ := uuid.Parse(req.ParentID)

	err := h.svc.CreateDirectoryNode(ctx, req.Name, parentId, user.ID)
	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusInternalServerError, "error creating directory node")
	}
	return c.JSON(http.StatusCreated, "directory created")
}

func (h Handler) Copy(
	c echo.Context,
) error {
	var req Move
	ctx := c.Request().Context()

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "invalid request body")
	}
	var user *auth.CustomClaims = c.Get("user").(*auth.CustomClaims)
	targetNodeId, _ := uuid.Parse(req.TargetNodeID)
	destParentId, _ := uuid.Parse(req.DestParentID)

	err := h.svc.Copy(ctx, targetNodeId, destParentId, user.ID)
	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusInternalServerError, "error performing copy operation")
	}
	return c.JSON(http.StatusAccepted, "copy successful")
}

func (h Handler) Move(
	c echo.Context,
) error {
	var req Move
	ctx := c.Request().Context()

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "invalid request body")
	}
	var user *auth.CustomClaims = c.Get("user").(*auth.CustomClaims)
	targetNodeId, _ := uuid.Parse(req.TargetNodeID)
	destParentId, _ := uuid.Parse(req.DestParentID)

	err := h.svc.Move(ctx, targetNodeId, destParentId, user.ID)
	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusInternalServerError, "error performing move operation")
	}
	return c.JSON(http.StatusAccepted, "move successful")
}

func (h Handler) Delete(
	c echo.Context,
) error {
	var req Delete
	ctx := c.Request().Context()

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, "invalid request body")
	}
	var user *auth.CustomClaims = c.Get("user").(*auth.CustomClaims)
	targetNodeId, _ := uuid.Parse(req.NodeID)
	
	err := h.svc.Delete(ctx, targetNodeId , user.ID)

	if err != nil {
		log.Println(err.Error())
		return c.JSON(http.StatusInternalServerError, "error deleting selected node")
	}

	return c.JSON(http.StatusAccepted, "deletion successful")
}
