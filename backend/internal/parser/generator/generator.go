// Package generator serializes a normalized model.Schema back into dialect
// DDL — the inverse of internal/parser. It is a thin dispatcher: each supported
// dialect has its own renderer that walks the schema and emits CREATE TYPE /
// CREATE TABLE statements designed to round-trip back through the matching
// parser to the same model.Schema (modulo documented limitations).
package generator

import (
	"fmt"
	"strings"

	"github.com/enricojoe/dgram/backend/internal/model"
)

// Generate renders schema as DDL for the given dialect. A non-nil error
// indicates an unknown dialect; rendering itself does not fail.
func Generate(d model.Dialect, schema model.Schema) (string, error) {
	switch d {
	case model.DialectPostgres:
		return generatePostgres(schema), nil
	case model.DialectMySQL:
		return generateMySQL(schema), nil
	default:
		return "", fmt.Errorf("unsupported dialect %q", d)
	}
}

// groupRefsByTable buckets foreign-key refs by their owning (from) table so a
// table's constraints can be emitted together, preserving order.
func groupRefsByTable(refs []model.Ref) map[string][]model.Ref {
	out := make(map[string][]model.Ref)
	for _, r := range refs {
		out[r.FromTable] = append(out[r.FromTable], r)
	}
	return out
}

// primaryKeyColumns returns the names of the table's primary-key columns in
// declaration order.
func primaryKeyColumns(t model.Table) []string {
	var cols []string
	for _, c := range t.Columns {
		if c.PK {
			cols = append(cols, c.Name)
		}
	}
	return cols
}

// renderEnumValues renders enum values as a comma-separated list of
// single-quoted literals, doubling embedded quotes.
func renderEnumValues(values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = "'" + strings.ReplaceAll(v, "'", "''") + "'"
	}
	return strings.Join(quoted, ", ")
}
