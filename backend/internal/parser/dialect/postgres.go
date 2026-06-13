package dialect

import (
	"fmt"
	"regexp"
	"strings"

	pgparser "github.com/auxten/postgresql-parser/pkg/sql/parser"
	"github.com/auxten/postgresql-parser/pkg/sql/sem/tree"

	"github.com/enricojoe/dgram/backend/internal/model"
)

// ParsePostgres parses Postgres DDL into a normalized model.Schema.
//
// It walks the CockroachDB-derived AST from auxten/postgresql-parser: each
// statement's .AST is type-switched, *tree.CreateTable nodes are mapped to
// model.Table by ranging over their .Defs (ColumnTableDef / table-level
// UniqueConstraintTableDef / ForeignKeyConstraintTableDef), and foreign keys —
// inline column References and table-level FK constraints alike — become one
// model.Ref per column pair.
//
// The underlying parser does not implement "CREATE TYPE ... AS ENUM", so those
// statements are extracted with a regexp before parsing and the rest of the
// DDL is parsed normally.
func ParsePostgres(ddl string) (model.Schema, []string, error) {
	var schema model.Schema
	var warnings []string

	// Hoist CREATE TYPE ... AS ENUM out first (unsupported by the parser), and
	// rewrite any column that USES an enum type to "text" — the parser rejects
	// unknown user-defined types, so this keeps such tables parseable. The
	// column's enum linkage is not preserved (see package docs / known limits).
	ddl, enums := extractPostgresEnums(ddl)
	schema.Enums = enums
	ddl = substitutePostgresEnumUsages(ddl, enums)

	stmts, err := pgparser.Parse(ddl)
	if err != nil {
		// Tolerant fallback: the parser fails the whole batch on the first
		// syntax error, so retry statement-by-statement and warn on the bad
		// ones instead of failing the request outright.
		var ok bool
		stmts, warnings, ok = parsePostgresPerStatement(ddl)
		if !ok {
			return model.Schema{}, nil, fmt.Errorf("postgres parse error: %w", err)
		}
	}

	for _, s := range stmts {
		ct, isCreate := s.AST.(*tree.CreateTable)
		if !isCreate {
			warnings = append(warnings, fmt.Sprintf("skipped unsupported statement: %s", s.AST.StatementTag()))
			continue
		}
		table, refs := mapPostgresTable(ct)
		schema.Tables = append(schema.Tables, table)
		schema.Refs = append(schema.Refs, refs...)
	}

	return schema, warnings, nil
}

// parsePostgresPerStatement splits ddl on semicolons and parses each fragment
// independently, collecting successfully-parsed statements and warning about
// the rest. ok is false when nothing at all could be parsed.
func parsePostgresPerStatement(ddl string) (stmts pgparser.Statements, warnings []string, ok bool) {
	for _, frag := range strings.Split(ddl, ";") {
		frag = strings.TrimSpace(frag)
		if frag == "" {
			continue
		}
		parsed, err := pgparser.Parse(frag)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("skipped unparseable statement: %v", err))
			continue
		}
		stmts = append(stmts, parsed...)
		ok = true
	}
	return stmts, warnings, ok
}

// mapPostgresTable converts a CreateTable node into a model.Table plus any
// foreign-key Refs it declares (inline or table-level).
func mapPostgresTable(ct *tree.CreateTable) (model.Table, []model.Ref) {
	table := model.Table{
		Schema: postgresSchemaName(&ct.Table),
		Name:   ct.Table.Table(),
	}
	var refs []model.Ref

	for _, def := range ct.Defs {
		switch d := def.(type) {
		case *tree.ColumnTableDef:
			col := mapPostgresColumn(d)
			table.Columns = append(table.Columns, col)
			// Inline column-level REFERENCES.
			if d.HasFKConstraint() {
				refs = append(refs, model.Ref{
					FromTable:  table.Name,
					FromColumn: col.Name,
					ToTable:    d.References.Table.Table(),
					ToColumn:   string(d.References.Col),
					OnDelete:   postgresRefAction(d.References.Actions.Delete),
					OnUpdate:   postgresRefAction(d.References.Actions.Update),
				})
			}
		case *tree.UniqueConstraintTableDef:
			// Table-level PRIMARY KEY / UNIQUE: flag the named columns.
			markPostgresKeyColumns(&table, d)
		case *tree.ForeignKeyConstraintTableDef:
			// Table-level FOREIGN KEY: one Ref per column pair.
			for i, from := range d.FromCols {
				to := ""
				if i < len(d.ToCols) {
					to = string(d.ToCols[i])
				}
				refs = append(refs, model.Ref{
					FromTable:  table.Name,
					FromColumn: string(from),
					ToTable:    d.Table.Table(),
					ToColumn:   to,
					OnDelete:   postgresRefAction(d.Actions.Delete),
					OnUpdate:   postgresRefAction(d.Actions.Update),
				})
			}
		}
	}

	return table, refs
}

// mapPostgresColumn maps a single column definition, normalizing the type and
// folding serial/bigserial into AutoInc.
func mapPostgresColumn(d *tree.ColumnTableDef) model.Column {
	col := model.Column{
		Name:    string(d.Name),
		Type:    normalizePostgresType(d),
		PK:      d.PrimaryKey.IsPrimaryKey,
		Unique:  d.Unique,
		AutoInc: d.IsSerial,
		// NotNull when explicitly NOT NULL, or implied by PRIMARY KEY.
		NotNull: d.Nullable.Nullability == tree.NotNull || d.PrimaryKey.IsPrimaryKey,
	}
	if d.HasDefaultExpr() {
		def := d.DefaultExpr.Expr.String()
		col.Default = &def
	}
	return col
}

// markPostgresKeyColumns sets PK (and NotNull) or Unique on the columns named
// by a table-level UNIQUE/PRIMARY KEY constraint.
func markPostgresKeyColumns(table *model.Table, d *tree.UniqueConstraintTableDef) {
	for _, elem := range d.Columns {
		name := string(elem.Column)
		for i := range table.Columns {
			if table.Columns[i].Name == name {
				if d.PrimaryKey {
					table.Columns[i].PK = true
					table.Columns[i].NotNull = true
				} else {
					table.Columns[i].Unique = true
				}
			}
		}
	}
}

// normalizePostgresType renders a column type to friendly Postgres text. Serial
// columns are normalized to "serial" (AutoInc carries the autoincrement
// semantics); CockroachDB type spellings are mapped back to common Postgres
// names where they differ.
func normalizePostgresType(d *tree.ColumnTableDef) string {
	if d.IsSerial {
		return "serial"
	}
	raw := d.Type.SQLString()
	if friendly, ok := postgresTypeAliases[strings.ToUpper(raw)]; ok {
		return friendly
	}
	return strings.ToLower(raw)
}

// postgresTypeAliases maps CockroachDB type spellings to common Postgres names.
var postgresTypeAliases = map[string]string{
	"INT8":   "bigint",
	"INT4":   "integer",
	"INT2":   "smallint",
	"INT":    "bigint",
	"STRING": "text",
	"BOOL":   "boolean",
	"FLOAT8": "double precision",
	"FLOAT4": "real",
}

// postgresRefAction renders a referential action, returning "" for NO ACTION.
func postgresRefAction(a tree.ReferenceAction) string {
	if a == tree.NoAction {
		return ""
	}
	return a.String()
}

// postgresSchemaName extracts the schema component of a (possibly qualified)
// table name. Parts are stored in reverse order: [object, schema, catalog].
func postgresSchemaName(tn *tree.TableName) string {
	on := tn.ToUnresolvedObjectName()
	if on.NumParts >= 2 {
		return on.Parts[1]
	}
	return ""
}

// enumRegexp matches "CREATE TYPE name AS ENUM ('a', 'b', ...)".
var enumRegexp = regexp.MustCompile(`(?is)CREATE\s+TYPE\s+([\w".]+)\s+AS\s+ENUM\s*\(([^)]*)\)\s*;?`)

// enumValueRegexp matches single-quoted enum values.
var enumValueRegexp = regexp.MustCompile(`'((?:[^']|'')*)'`)

// extractPostgresEnums pulls CREATE TYPE ... AS ENUM definitions out of ddl
// (the parser cannot handle them) and returns the remaining DDL plus the parsed
// enums.
func extractPostgresEnums(ddl string) (string, []model.Enum) {
	matches := enumRegexp.FindAllStringSubmatch(ddl, -1)
	if len(matches) == 0 {
		return ddl, nil
	}

	enums := make([]model.Enum, 0, len(matches))
	for _, m := range matches {
		enum := model.Enum{Name: unquotePostgresIdent(m[1])}
		for _, v := range enumValueRegexp.FindAllStringSubmatch(m[2], -1) {
			enum.Values = append(enum.Values, strings.ReplaceAll(v[1], "''", "'"))
		}
		enums = append(enums, enum)
	}

	return enumRegexp.ReplaceAllString(ddl, ""), enums
}

// substitutePostgresEnumUsages replaces whole-word references to extracted enum
// type names with "text" so that columns typed as a user-defined enum (which
// the parser cannot resolve) still parse.
func substitutePostgresEnumUsages(ddl string, enums []model.Enum) string {
	for _, e := range enums {
		if e.Name == "" {
			continue
		}
		re := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(e.Name) + `\b`)
		ddl = re.ReplaceAllString(ddl, "text")
	}
	return ddl
}

// unquotePostgresIdent strips schema qualification and double-quotes from an
// identifier, returning just the object name.
func unquotePostgresIdent(ident string) string {
	parts := strings.Split(ident, ".")
	name := parts[len(parts)-1]
	return strings.Trim(name, `"`)
}
