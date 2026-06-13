package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/enricojoe/dgram/backend/internal/model"
	"github.com/enricojoe/dgram/backend/internal/service"
	"github.com/enricojoe/dgram/backend/internal/util"
)

// GenerateController exposes schema-to-DDL generation.
type GenerateController struct {
	schemas *service.SchemaService
}

func NewGenerateController(schemas *service.SchemaService) *GenerateController {
	return &GenerateController{schemas: schemas}
}

// Register mounts the generate routes on the given router group.
func (g *GenerateController) Register(r *gin.RouterGroup) {
	r.POST("/generate", g.generate)
}

// generateRequest is the POST /generate request body.
type generateRequest struct {
	Dialect model.Dialect `json:"dialect"`
	Schema  model.Schema  `json:"schema"`
}

// generateResponse is the POST /generate response body.
type generateResponse struct {
	DDL string `json:"ddl"`
}

func (g *GenerateController) generate(c *gin.Context) {
	var req generateRequest
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

	ddl, err := g.schemas.Generate(req.Dialect, req.Schema)
	if err != nil {
		util.Fail(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	util.OK(c, generateResponse{DDL: ddl})
}
