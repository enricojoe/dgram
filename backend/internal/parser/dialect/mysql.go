package dialect

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/format"
	"github.com/pingcap/tidb/pkg/parser/mysql"
	"github.com/pingcap/tidb/pkg/parser/types"
	// The standalone TiDB parser needs a value-expression driver registered at
	// init time; test_driver is the canonical one for SQL-text parsing.
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"

	"github.com/enricojoe/dgram/backend/internal/model"
)

// ParseMySQL parses MySQL DDL into a normalized model.Schema.
//
// It uses parser.New().Parse(...) to obtain TiDB AST nodes, type-switches each
// to *ast.CreateTableStmt, then maps .Cols (with their .Options:
// PrimaryKey/NotNull/AutoIncrement/UniqKey/DefaultValue/Reference) and
// .Constraints (table-level PRIMARY KEY / UNIQUE / FOREIGN KEY). Each FK column
// pair becomes one model.Ref, and inline ENUM(...) columns are hoisted into a
// named model.Enum.
func ParseMySQL(ddl string) (model.Schema, []string, error) {
	var schema model.Schema
	var warnings []string

	p := parser.New()
	stmts, warns, err := p.Parse(ddl, "", "")
	if err != nil {
		return model.Schema{}, nil, fmt.Errorf("mysql parse error: %w", err)
	}
	for _, w := range warns {
		warnings = append(warnings, w.Error())
	}

	for _, stmt := range stmts {
		ct, isCreate := stmt.(*ast.CreateTableStmt)
		if !isCreate {
			warnings = append(warnings, fmt.Sprintf("skipped unsupported statement: %T", stmt))
			continue
		}
		table, refs, enums := mapMySQLTable(ct)
		schema.Tables = append(schema.Tables, table)
		schema.Refs = append(schema.Refs, refs...)
		schema.Enums = append(schema.Enums, enums...)
	}

	return schema, warnings, nil
}

// mapMySQLTable converts a CreateTableStmt into a model.Table plus the Refs and
// hoisted enums it implies.
func mapMySQLTable(ct *ast.CreateTableStmt) (model.Table, []model.Ref, []model.Enum) {
	table := model.Table{
		Schema: ct.Table.Schema.O,
		Name:   ct.Table.Name.O,
	}
	var refs []model.Ref
	var enums []model.Enum

	for _, c := range ct.Cols {
		col, enum := mapMySQLColumn(table.Name, c)
		table.Columns = append(table.Columns, col)
		if enum != nil {
			enums = append(enums, *enum)
		}
		// Inline column-level REFERENCES.
		for _, opt := range c.Options {
			if opt.Tp == ast.ColumnOptionReference && opt.Refer != nil {
				refs = append(refs, mySQLColumnRefs(table.Name, col.Name, opt.Refer)...)
			}
		}
	}

	for _, cons := range ct.Constraints {
		switch cons.Tp {
		case ast.ConstraintPrimaryKey:
			markMySQLKeyColumns(&table, cons, true)
		case ast.ConstraintUniq, ast.ConstraintUniqKey, ast.ConstraintUniqIndex:
			markMySQLKeyColumns(&table, cons, false)
		case ast.ConstraintForeignKey:
			refs = append(refs, mapMySQLForeignKey(table.Name, cons)...)
		}
	}

	return table, refs, enums
}

// mapMySQLColumn maps a column definition, folding AUTO_INCREMENT into AutoInc
// and hoisting inline ENUM(...) into a named enum (returned non-nil when found).
func mapMySQLColumn(tableName string, c *ast.ColumnDef) (model.Column, *model.Enum) {
	col := model.Column{
		Name: c.Name.Name.O,
		Type: normalizeMySQLType(c.Tp),
	}

	var enum *model.Enum
	if c.Tp.GetType() == mysql.TypeEnum {
		enumName := tableName + "_" + col.Name
		enum = &model.Enum{Name: enumName, Values: append([]string(nil), c.Tp.GetElems()...)}
		col.Type = enumName
	}

	for _, opt := range c.Options {
		switch opt.Tp {
		case ast.ColumnOptionPrimaryKey:
			col.PK = true
			col.NotNull = true
		case ast.ColumnOptionNotNull:
			col.NotNull = true
		case ast.ColumnOptionAutoIncrement:
			col.AutoInc = true
		case ast.ColumnOptionUniqKey:
			col.Unique = true
		case ast.ColumnOptionDefaultValue:
			if def := renderMySQLExpr(opt.Expr); def != "" {
				col.Default = &def
			}
		}
	}

	return col, enum
}

// mapMySQLForeignKey expands a table-level FOREIGN KEY constraint into one Ref
// per (from, to) column pair.
func mapMySQLForeignKey(tableName string, cons *ast.Constraint) []model.Ref {
	if cons.Refer == nil {
		return nil
	}
	return mySQLRefs(tableName, cons.Keys, cons.Refer)
}

// mySQLColumnRefs builds Refs for an inline column-level REFERENCES clause.
func mySQLColumnRefs(tableName, colName string, refer *ast.ReferenceDef) []model.Ref {
	toCol := ""
	if len(refer.IndexPartSpecifications) > 0 {
		toCol = refer.IndexPartSpecifications[0].Column.Name.O
	}
	return []model.Ref{{
		FromTable:  tableName,
		FromColumn: colName,
		ToTable:    refer.Table.Name.O,
		ToColumn:   toCol,
		OnDelete:   mySQLRefOption(refer.OnDelete),
		OnUpdate:   mySQLRefOptionUpdate(refer.OnUpdate),
	}}
}

// mySQLRefs pairs local key columns with referenced columns positionally.
func mySQLRefs(tableName string, keys []*ast.IndexPartSpecification, refer *ast.ReferenceDef) []model.Ref {
	refs := make([]model.Ref, 0, len(keys))
	for i, k := range keys {
		toCol := ""
		if i < len(refer.IndexPartSpecifications) {
			toCol = refer.IndexPartSpecifications[i].Column.Name.O
		}
		refs = append(refs, model.Ref{
			FromTable:  tableName,
			FromColumn: k.Column.Name.O,
			ToTable:    refer.Table.Name.O,
			ToColumn:   toCol,
			OnDelete:   mySQLRefOption(refer.OnDelete),
			OnUpdate:   mySQLRefOptionUpdate(refer.OnUpdate),
		})
	}
	return refs
}

// markMySQLKeyColumns flags the columns named by a table-level PRIMARY KEY or
// UNIQUE constraint.
func markMySQLKeyColumns(table *model.Table, cons *ast.Constraint, primary bool) {
	for _, k := range cons.Keys {
		name := k.Column.Name.O
		for i := range table.Columns {
			if table.Columns[i].Name == name {
				if primary {
					table.Columns[i].PK = true
					table.Columns[i].NotNull = true
				} else {
					table.Columns[i].Unique = true
				}
			}
		}
	}
}

// intTypeWidth matches a trailing display width on integer types, e.g. "(20)".
var intTypeWidth = regexp.MustCompile(`\(\d+\)$`)

// normalizeMySQLType renders a field type to a normalized string, dropping the
// cosmetic display width from integer types (e.g. "bigint(20)" -> "bigint").
func normalizeMySQLType(ft *types.FieldType) string {
	s := ft.CompactStr()
	switch ft.GetType() {
	case mysql.TypeTiny, mysql.TypeShort, mysql.TypeInt24, mysql.TypeLong, mysql.TypeLonglong:
		s = intTypeWidth.ReplaceAllString(s, "")
	}
	return s
}

// renderMySQLExpr restores an expression node to its SQL text (used for
// DEFAULT values). Returns "" if rendering fails.
func renderMySQLExpr(expr ast.ExprNode) string {
	if expr == nil {
		return ""
	}
	var sb strings.Builder
	// Omit the charset introducer (e.g. _utf8mb4'x') from string literals.
	flags := format.RestoreStringSingleQuotes | format.RestoreKeyWordUppercase | format.RestoreNameBackQuotes | format.RestoreStringWithoutCharset
	ctx := format.NewRestoreCtx(flags, &sb)
	if err := expr.Restore(ctx); err != nil {
		return ""
	}
	return sb.String()
}

// mySQLRefOption renders an ON DELETE action, returning "" when unset/NO ACTION.
func mySQLRefOption(o *ast.OnDeleteOpt) string {
	if o == nil {
		return ""
	}
	return normalizeRefOption(o.ReferOpt.String())
}

// mySQLRefOptionUpdate renders an ON UPDATE action.
func mySQLRefOptionUpdate(o *ast.OnUpdateOpt) string {
	if o == nil {
		return ""
	}
	return normalizeRefOption(o.ReferOpt.String())
}

// normalizeRefOption treats "NO ACTION" (and empties) as unset.
func normalizeRefOption(s string) string {
	if s == "" || strings.EqualFold(s, "NO ACTION") {
		return ""
	}
	return s
}
