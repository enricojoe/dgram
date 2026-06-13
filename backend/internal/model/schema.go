package model

// schema.go defines the NORMALIZED schema model — the single shape that every
// dialect parser produces and the generator consumes. The frontend mirrors
// these types in TypeScript. Treat this as a frozen contract: additive changes
// are fine, but renames/removals ripple through parser, generator, and UI.

// Dialect identifies a supported SQL dialect.
type Dialect string

const (
	DialectPostgres Dialect = "postgres"
	DialectMySQL    Dialect = "mysql"
)

// Schema is the full parsed result: tables and the relationships between them.
type Schema struct {
	Tables []Table `json:"tables"`
	Refs   []Ref   `json:"refs"`
	Enums  []Enum  `json:"enums"`
}

// Table is a single CREATE TABLE definition.
type Table struct {
	// Schema is the SQL schema/namespace (e.g. "public"). Empty when unqualified.
	Schema  string   `json:"schema,omitempty"`
	Name    string   `json:"name"`
	Columns []Column `json:"columns"`
	Indexes []Index  `json:"indexes,omitempty"`
	// Note is an optional human comment surfaced in the diagram.
	Note string `json:"note,omitempty"`
}

// Column is one column within a table.
type Column struct {
	Name string `json:"name"`
	// Type is the normalized type text, e.g. "varchar(255)", "int", "uuid".
	Type    string `json:"type"`
	PK      bool   `json:"pk,omitempty"`
	NotNull bool   `json:"notNull,omitempty"`
	Unique  bool   `json:"unique,omitempty"`
	AutoInc bool   `json:"autoInc,omitempty"`
	// Default is the raw default expression, or nil when none.
	Default *string `json:"default,omitempty"`
	Note    string  `json:"note,omitempty"`
}

// Ref is a foreign-key relationship (one FK column → one referenced column).
type Ref struct {
	FromTable  string `json:"fromTable"`
	FromColumn string `json:"fromColumn"`
	ToTable    string `json:"toTable"`
	ToColumn   string `json:"toColumn"`
	// OnDelete / OnUpdate are referential actions, e.g. "CASCADE". Empty if unset.
	OnDelete string `json:"onDelete,omitempty"`
	OnUpdate string `json:"onUpdate,omitempty"`
}

// Index is a (possibly unique) index on one or more columns.
type Index struct {
	Name    string   `json:"name,omitempty"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique,omitempty"`
}

// Enum is a named enumerated type (Postgres CREATE TYPE ... AS ENUM, or MySQL
// inline enum hoisted to a named type).
type Enum struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}
