package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/enricojoe/dgram/backend/internal/middleware"
	"github.com/enricojoe/dgram/backend/internal/service"
	"github.com/enricojoe/dgram/backend/internal/util"
)

// DiagramController exposes CRUD over the authenticated user's diagrams.
type DiagramController struct {
	diagrams *service.DiagramService
}

func NewDiagramController(diagrams *service.DiagramService) *DiagramController {
	return &DiagramController{diagrams: diagrams}
}

// Register mounts the diagram routes; the group is expected to already require
// authentication.
func (d *DiagramController) Register(r *gin.RouterGroup) {
	g := r.Group("/diagrams")
	g.GET("", d.list)
	g.POST("", d.create)
	g.GET("/:id", d.get)
	g.PUT("/:id", d.update)
	g.DELETE("/:id", d.delete)
	g.POST("/:id/share", d.share)
	g.DELETE("/:id/share", d.unshare)
}

// RegisterPublic mounts the unauthenticated share-view route. The group is
// expected to NOT require authentication.
func (d *DiagramController) RegisterPublic(r *gin.RouterGroup) {
	r.GET("/share/:shareId", d.getShared)
}

type createDiagramRequest struct {
	Name    string          `json:"name"`
	Dialect string          `json:"dialect"`
	DDL     string          `json:"ddl"`
	Layout  json.RawMessage `json:"layout"`
}

type updateDiagramRequest struct {
	Name    *string         `json:"name"`
	Dialect *string         `json:"dialect"`
	DDL     *string         `json:"ddl"`
	Layout  json.RawMessage `json:"layout"`
}

func (d *DiagramController) list(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		util.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	items, err := d.diagrams.List(userID)
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "could not list diagrams")
		return
	}
	util.OK(c, items)
}

func (d *DiagramController) create(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		util.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	var req createDiagramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if strings.TrimSpace(req.Name) == "" {
		util.Fail(c, http.StatusBadRequest, "name must not be empty")
		return
	}
	if strings.TrimSpace(req.Dialect) == "" {
		util.Fail(c, http.StatusBadRequest, "dialect must not be empty")
		return
	}

	diagram, err := d.diagrams.Create(userID, req.Name, req.Dialect, req.DDL, req.Layout)
	if err != nil {
		util.Fail(c, http.StatusInternalServerError, "could not create diagram")
		return
	}
	util.Created(c, diagram)
}

func (d *DiagramController) get(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		util.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}

	diagram, err := d.diagrams.Get(id, userID)
	if err != nil {
		if errors.Is(err, service.ErrDiagramNotFound) {
			util.Fail(c, http.StatusNotFound, "diagram not found")
			return
		}
		util.Fail(c, http.StatusInternalServerError, "could not load diagram")
		return
	}
	util.OK(c, diagram)
}

func (d *DiagramController) update(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		util.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req updateDiagramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	diagram, err := d.diagrams.Update(id, userID, req.Name, req.Dialect, req.DDL, req.Layout)
	if err != nil {
		if errors.Is(err, service.ErrDiagramNotFound) {
			util.Fail(c, http.StatusNotFound, "diagram not found")
			return
		}
		util.Fail(c, http.StatusInternalServerError, "could not update diagram")
		return
	}
	util.OK(c, diagram)
}

func (d *DiagramController) delete(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		util.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := d.diagrams.Delete(id, userID); err != nil {
		if errors.Is(err, service.ErrDiagramNotFound) {
			util.Fail(c, http.StatusNotFound, "diagram not found")
			return
		}
		util.Fail(c, http.StatusInternalServerError, "could not delete diagram")
		return
	}
	c.Status(http.StatusNoContent)
}

// share enables public sharing for the diagram and returns its share token.
// Idempotent: an already-shared diagram returns its existing token.
func (d *DiagramController) share(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		util.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}

	shareID, err := d.diagrams.EnableShare(id, userID)
	if err != nil {
		if errors.Is(err, service.ErrDiagramNotFound) {
			util.Fail(c, http.StatusNotFound, "diagram not found")
			return
		}
		util.Fail(c, http.StatusInternalServerError, "could not share diagram")
		return
	}
	util.OK(c, gin.H{"shareId": shareID, "isPublic": true})
}

// unshare disables public access while preserving the share token.
func (d *DiagramController) unshare(c *gin.Context) {
	userID, ok := middleware.UserID(c)
	if !ok {
		util.Fail(c, http.StatusUnauthorized, "unauthorized")
		return
	}
	id, ok := parseID(c)
	if !ok {
		return
	}

	if err := d.diagrams.DisableShare(id, userID); err != nil {
		if errors.Is(err, service.ErrDiagramNotFound) {
			util.Fail(c, http.StatusNotFound, "diagram not found")
			return
		}
		util.Fail(c, http.StatusInternalServerError, "could not unshare diagram")
		return
	}
	util.OK(c, gin.H{"isPublic": false})
}

// getShared serves the public read-only view of a shared diagram. Unknown or
// non-public tokens return 404 without leaking whether the diagram exists.
func (d *DiagramController) getShared(c *gin.Context) {
	shareID := c.Param("shareId")

	diagram, err := d.diagrams.GetShared(shareID)
	if err != nil {
		util.Fail(c, http.StatusNotFound, "not found")
		return
	}
	util.OK(c, diagram)
}

// parseID extracts the :id path parameter, writing a 404 and returning false
// when it is not a valid integer (an invalid id can never match a real row).
func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		util.Fail(c, http.StatusNotFound, "diagram not found")
		return 0, false
	}
	return id, true
}
