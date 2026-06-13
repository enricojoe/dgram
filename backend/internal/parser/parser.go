// Package parser turns raw DDL text into the normalized model.Schema shape.
//
// It is a thin dispatcher: each supported dialect has its own mapper under
// internal/parser/dialect that walks that dialect's AST and emits the same
// model.Schema. Parsing is tolerant — statements that cannot be mapped are
// reported as warnings rather than failing the whole request.
package parser

import (
	"fmt"

	"github.com/enricojoe/dgram/backend/internal/model"
	"github.com/enricojoe/dgram/backend/internal/parser/dialect"
)

// Parse parses ddl for the given dialect and returns the normalized schema plus
// a slice of non-fatal warning strings (unsupported/skipped statements). A
// non-nil error indicates a fatal failure (e.g. an unparseable input or an
// unknown dialect) and the returned schema should be ignored.
func Parse(d model.Dialect, ddl string) (model.Schema, []string, error) {
	switch d {
	case model.DialectPostgres:
		return dialect.ParsePostgres(ddl)
	case model.DialectMySQL:
		return dialect.ParseMySQL(ddl)
	default:
		return model.Schema{}, nil, fmt.Errorf("unsupported dialect %q", d)
	}
}
