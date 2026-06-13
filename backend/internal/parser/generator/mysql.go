package generator

import (
	"strings"

	"github.com/enricojoe/dgram/backend/internal/model"
)

// generateMySQL renders schema as MySQL DDL.
//
// MySQL has no standalone enum type, so enums are emitted inline as ENUM(...) on
// the columns that reference them (a column whose Type equals an enum name); the
// parser hoists those back into named enums. Identifiers are backtick-quoted,
// AutoInc columns use AUTO_INCREMENT, a sole primary key is inlined, and foreign
// keys are emitted as table-level constraints grouped by their owning table.
func generateMySQL(schema model.Schema) string {
	var sb strings.Builder

	enumValues := make(map[string][]string, len(schema.Enums))
	for _, e := range schema.Enums {
		enumValues[e.Name] = e.Values
	}

	refsByTable := groupRefsByTable(schema.Refs)

	for ti, t := range schema.Tables {
		if ti > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("CREATE TABLE ")
		sb.WriteString(mysqlTableName(t))
		sb.WriteString(" (\n")

		var lines []string
		pkCols := primaryKeyColumns(t)
		inlinePK := len(pkCols) == 1

		for _, col := range t.Columns {
			lines = append(lines, "    "+mysqlColumn(col, inlinePK, enumValues))
		}
		if !inlinePK && len(pkCols) > 1 {
			lines = append(lines, "    PRIMARY KEY ("+strings.Join(mysqlQuoteAll(pkCols), ", ")+")")
		}
		for _, r := range refsByTable[t.Name] {
			lines = append(lines, "    "+mysqlForeignKey(r))
		}

		sb.WriteString(strings.Join(lines, ",\n"))
		sb.WriteString("\n);\n")
	}

	return sb.String()
}

// mysqlTableName renders a (possibly schema-qualified) backtick-quoted name.
func mysqlTableName(t model.Table) string {
	if t.Schema != "" {
		return mysqlQuote(t.Schema) + "." + mysqlQuote(t.Name)
	}
	return mysqlQuote(t.Name)
}

// mysqlColumn renders a single column definition line.
func mysqlColumn(col model.Column, inlinePK bool, enumValues map[string][]string) string {
	parts := []string{mysqlQuote(col.Name), mysqlColumnType(col, enumValues)}

	if col.NotNull {
		parts = append(parts, "NOT NULL")
	}
	if col.AutoInc {
		parts = append(parts, "AUTO_INCREMENT")
	}
	if col.Unique {
		parts = append(parts, "UNIQUE")
	}
	if !col.AutoInc && col.Default != nil {
		parts = append(parts, "DEFAULT "+*col.Default)
	}
	if inlinePK && col.PK {
		parts = append(parts, "PRIMARY KEY")
	}

	return strings.Join(parts, " ")
}

// mysqlColumnType resolves a column's rendered type: an inline ENUM(...) when
// the type names an enum, an integer type for AutoInc columns (MySQL has no
// serial), or the stored type otherwise.
func mysqlColumnType(col model.Column, enumValues map[string][]string) string {
	if vals, ok := enumValues[col.Type]; ok {
		return "enum(" + renderEnumValues(vals) + ")"
	}
	if col.AutoInc {
		switch col.Type {
		case "serial":
			return "int"
		case "bigserial":
			return "bigint"
		default:
			return col.Type
		}
	}
	return col.Type
}

// mysqlForeignKey renders a table-level FOREIGN KEY constraint.
func mysqlForeignKey(r model.Ref) string {
	var sb strings.Builder
	sb.WriteString("FOREIGN KEY (")
	sb.WriteString(mysqlQuote(r.FromColumn))
	sb.WriteString(") REFERENCES ")
	sb.WriteString(mysqlQuote(r.ToTable))
	sb.WriteString(" (")
	sb.WriteString(mysqlQuote(r.ToColumn))
	sb.WriteString(")")
	if r.OnDelete != "" {
		sb.WriteString(" ON DELETE " + r.OnDelete)
	}
	if r.OnUpdate != "" {
		sb.WriteString(" ON UPDATE " + r.OnUpdate)
	}
	return sb.String()
}

// mysqlQuote backtick-quotes an identifier, escaping embedded backticks.
func mysqlQuote(ident string) string {
	return "`" + strings.ReplaceAll(ident, "`", "``") + "`"
}

// mysqlQuoteAll backtick-quotes each identifier in the slice.
func mysqlQuoteAll(idents []string) []string {
	out := make([]string, len(idents))
	for i, id := range idents {
		out[i] = mysqlQuote(id)
	}
	return out
}
