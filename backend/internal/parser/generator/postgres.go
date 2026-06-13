package generator

import (
	"strings"

	"github.com/enricojoe/dgram/backend/internal/model"
)

// generatePostgres renders schema as PostgreSQL DDL.
//
// Enums are emitted first as CREATE TYPE ... AS ENUM (the parser hoists these
// back out before parsing), then one CREATE TABLE per table. Columns flagged
// AutoInc are emitted as "serial" (matching how the parser normalizes serial
// columns) and never carry a DEFAULT. A single primary-key column is inlined as
// PRIMARY KEY; composite keys become a table-level constraint. Foreign keys are
// grouped by their owning table and emitted as table-level constraints.
func generatePostgres(schema model.Schema) string {
	var sb strings.Builder

	for _, e := range schema.Enums {
		sb.WriteString("CREATE TYPE ")
		sb.WriteString(e.Name)
		sb.WriteString(" AS ENUM (")
		sb.WriteString(renderEnumValues(e.Values))
		sb.WriteString(");\n")
	}
	if len(schema.Enums) > 0 {
		sb.WriteString("\n")
	}

	refsByTable := groupRefsByTable(schema.Refs)

	for ti, t := range schema.Tables {
		if ti > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("CREATE TABLE ")
		sb.WriteString(postgresTableName(t))
		sb.WriteString(" (\n")

		var lines []string
		pkCols := primaryKeyColumns(t)
		inlinePK := len(pkCols) == 1

		for _, col := range t.Columns {
			lines = append(lines, "    "+postgresColumn(col, inlinePK))
		}
		if !inlinePK && len(pkCols) > 1 {
			lines = append(lines, "    PRIMARY KEY ("+strings.Join(pkCols, ", ")+")")
		}
		for _, r := range refsByTable[t.Name] {
			lines = append(lines, "    "+postgresForeignKey(r))
		}

		sb.WriteString(strings.Join(lines, ",\n"))
		sb.WriteString("\n);\n")
	}

	return sb.String()
}

// postgresTableName renders a (possibly schema-qualified) table name.
func postgresTableName(t model.Table) string {
	if t.Schema != "" {
		return t.Schema + "." + t.Name
	}
	return t.Name
}

// postgresColumn renders a single column definition line. When inlinePK is true
// and the column is the table's sole primary key, PRIMARY KEY is appended.
func postgresColumn(col model.Column, inlinePK bool) string {
	parts := []string{col.Name, postgresColumnType(col)}

	pkInlined := inlinePK && col.PK
	if pkInlined {
		parts = append(parts, "PRIMARY KEY")
	} else if col.NotNull {
		// PRIMARY KEY already implies NOT NULL, so only emit it otherwise.
		parts = append(parts, "NOT NULL")
	}
	if col.Unique {
		parts = append(parts, "UNIQUE")
	}
	// AutoInc (serial) columns own their default; don't also emit one.
	if !col.AutoInc && col.Default != nil {
		parts = append(parts, "DEFAULT "+*col.Default)
	}

	return strings.Join(parts, " ")
}

// postgresColumnType returns "serial" for AutoInc columns (so the parser folds
// it back into AutoInc) and the stored type otherwise.
func postgresColumnType(col model.Column) string {
	if col.AutoInc {
		return "serial"
	}
	return col.Type
}

// postgresForeignKey renders a table-level FOREIGN KEY constraint.
func postgresForeignKey(r model.Ref) string {
	var sb strings.Builder
	sb.WriteString("FOREIGN KEY (")
	sb.WriteString(r.FromColumn)
	sb.WriteString(") REFERENCES ")
	sb.WriteString(r.ToTable)
	sb.WriteString("(")
	sb.WriteString(r.ToColumn)
	sb.WriteString(")")
	if r.OnDelete != "" {
		sb.WriteString(" ON DELETE " + r.OnDelete)
	}
	if r.OnUpdate != "" {
		sb.WriteString(" ON UPDATE " + r.OnUpdate)
	}
	return sb.String()
}
