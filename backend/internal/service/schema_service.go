// Package service holds the application's business logic, sitting between the
// HTTP controllers and the lower-level parser/generator engines.
package service

import (
	"github.com/enricojoe/dgram/backend/internal/model"
	"github.com/enricojoe/dgram/backend/internal/parser"
	"github.com/enricojoe/dgram/backend/internal/parser/generator"
)

// SchemaService coordinates schema parsing (and, later, diagram generation).
type SchemaService struct{}

// NewSchemaService constructs a SchemaService.
func NewSchemaService() *SchemaService {
	return &SchemaService{}
}

// Parse parses ddl for the given dialect into a normalized schema, returning
// any non-fatal warnings alongside it. A non-nil error is a fatal parse
// failure.
func (s *SchemaService) Parse(dialect model.Dialect, ddl string) (model.Schema, []string, error) {
	return parser.Parse(dialect, ddl)
}

// Generate serializes schema into DDL for the given dialect. A non-nil error
// indicates an unknown dialect.
func (s *SchemaService) Generate(dialect model.Dialect, schema model.Schema) (string, error) {
	return generator.Generate(dialect, schema)
}
