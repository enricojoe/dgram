package controller

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/enricojoe/dgram/backend/internal/model"
	"github.com/enricojoe/dgram/backend/internal/service"
	"github.com/enricojoe/dgram/backend/internal/util"
)

// ParseController exposes DDL parsing.
type ParseController struct {
	schemas *service.SchemaService
}

func NewParseController(schemas *service.SchemaService) *ParseController {
	return &ParseController{schemas: schemas}
}

// Register mounts the parse routes on the given router group.
func (p *ParseController) Register(r *gin.RouterGroup) {
	r.POST("/parse", p.parse)
}

// parseRequest is the POST /parse request body.
type parseRequest struct {
	Dialect model.Dialect `json:"dialect"`
	DDL     string        `json:"ddl"`
}

// parseResponse is the POST /parse response body.
type parseResponse struct {
	Schema   model.Schema `json:"schema"`
	Warnings []string     `json:"warnings"`
}

func (p *ParseController) parse(c *gin.Context) {
	var req parseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.Fail(c, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	switch req.Dialect {
	case model.DialectPostgres, model.DialectMySQL:
		// supported
	default:
		util.Fail(c, http.StatusBadRequest, `dialect must be "postgres" or "mysql"`)
		return
	}

	if strings.TrimSpace(req.DDL) == "" {
		util.Fail(c, http.StatusBadRequest, "ddl must not be empty")
		return
	}

	schema, warnings, err := p.schemas.Parse(req.Dialect, req.DDL)
	if err != nil {
		util.Fail(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// Always emit a (possibly empty) warnings array rather than null.
	if warnings == nil {
		warnings = []string{}
	}
	util.OK(c, parseResponse{Schema: schema, Warnings: warnings})
}
