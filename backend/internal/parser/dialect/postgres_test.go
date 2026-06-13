package dialect

import (
	"testing"

	"github.com/enricojoe/dgram/backend/internal/model"
)

// findTable returns the table with the given name, or nil.
func findTable(s model.Schema, name string) *model.Table {
	for i := range s.Tables {
		if s.Tables[i].Name == name {
			return &s.Tables[i]
		}
	}
	return nil
}

// findColumn returns the column with the given name, or nil.
func findColumn(t *model.Table, name string) *model.Column {
	if t == nil {
		return nil
	}
	for i := range t.Columns {
		if t.Columns[i].Name == name {
			return &t.Columns[i]
		}
	}
	return nil
}

// findRef returns the first Ref matching from-table/from-column, or nil.
func findRef(s model.Schema, fromTable, fromCol string) *model.Ref {
	for i := range s.Refs {
		if s.Refs[i].FromTable == fromTable && s.Refs[i].FromColumn == fromCol {
			return &s.Refs[i]
		}
	}
	return nil
}

const postgresDDL = `
CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy');

CREATE TABLE users (
    id serial PRIMARY KEY,
    email varchar(255) UNIQUE NOT NULL,
    state mood
);

CREATE TABLE posts (
    id serial PRIMARY KEY,
    user_id int REFERENCES users(id) ON DELETE CASCADE,
    title varchar(100) NOT NULL
);
`

func TestParsePostgres(t *testing.T) {
	schema, warnings, err := ParsePostgres(postgresDDL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Logf("warnings: %v", warnings)
	}

	if len(schema.Tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(schema.Tables))
	}

	// users table.
	users := findTable(schema, "users")
	if users == nil {
		t.Fatal("users table not found")
	}
	if len(users.Columns) != 3 {
		t.Fatalf("users: expected 3 columns, got %d", len(users.Columns))
	}
	id := findColumn(users, "id")
	if id == nil || !id.PK || !id.AutoInc || !id.NotNull {
		t.Errorf("users.id: want pk+autoInc+notNull, got %+v", id)
	}
	email := findColumn(users, "email")
	if email == nil || !email.Unique || !email.NotNull {
		t.Errorf("users.email: want unique+notNull, got %+v", email)
	}
	if email != nil && email.Type != "varchar(255)" {
		t.Errorf("users.email: want type varchar(255), got %q", email.Type)
	}
	// Enum-typed column: the parser can't resolve user-defined types, so the
	// type is rewritten to "text" (documented limitation). The column survives.
	state := findColumn(users, "state")
	if state == nil {
		t.Error("users.state column missing")
	} else if state.Type != "text" {
		t.Errorf("users.state: want type text, got %q", state.Type)
	}

	// posts table + FK.
	posts := findTable(schema, "posts")
	if posts == nil {
		t.Fatal("posts table not found")
	}
	ref := findRef(schema, "posts", "user_id")
	if ref == nil {
		t.Fatal("posts.user_id FK not found")
	}
	if ref.ToTable != "users" || ref.ToColumn != "id" {
		t.Errorf("FK target: want users.id, got %s.%s", ref.ToTable, ref.ToColumn)
	}
	if ref.OnDelete != "CASCADE" {
		t.Errorf("FK onDelete: want CASCADE, got %q", ref.OnDelete)
	}

	// enum.
	if len(schema.Enums) != 1 {
		t.Fatalf("expected 1 enum, got %d", len(schema.Enums))
	}
	if schema.Enums[0].Name != "mood" {
		t.Errorf("enum name: want mood, got %q", schema.Enums[0].Name)
	}
	if got := schema.Enums[0].Values; len(got) != 3 || got[0] != "sad" || got[2] != "happy" {
		t.Errorf("enum values: want [sad ok happy], got %v", got)
	}
}

func TestParsePostgresTableLevelFK(t *testing.T) {
	const ddl = `
CREATE TABLE app.orders (
    id serial PRIMARY KEY,
    customer_id int NOT NULL,
    CONSTRAINT fk_customer FOREIGN KEY (customer_id) REFERENCES customers(id) ON DELETE RESTRICT ON UPDATE CASCADE
);`
	schema, _, err := ParsePostgres(ddl)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	orders := findTable(schema, "orders")
	if orders == nil {
		t.Fatal("orders table not found")
	}
	if orders.Schema != "app" {
		t.Errorf("orders schema: want app, got %q", orders.Schema)
	}
	ref := findRef(schema, "orders", "customer_id")
	if ref == nil {
		t.Fatal("FK not found")
	}
	if ref.ToTable != "customers" || ref.ToColumn != "id" {
		t.Errorf("FK target: want customers.id, got %s.%s", ref.ToTable, ref.ToColumn)
	}
	if ref.OnDelete != "RESTRICT" || ref.OnUpdate != "CASCADE" {
		t.Errorf("FK actions: want onDelete=RESTRICT onUpdate=CASCADE, got %q/%q", ref.OnDelete, ref.OnUpdate)
	}
}

func TestParsePostgresTolerant(t *testing.T) {
	// One bad statement should be warned about, not fatal.
	const ddl = `
CREATE TABLE good (id serial PRIMARY KEY);
THIS IS NOT VALID SQL;
CREATE TABLE alsogood (id serial PRIMARY KEY);`
	schema, warnings, err := ParsePostgres(ddl)
	if err != nil {
		t.Fatalf("unexpected fatal error: %v", err)
	}
	if len(schema.Tables) != 2 {
		t.Errorf("expected 2 tables despite bad statement, got %d", len(schema.Tables))
	}
	if len(warnings) == 0 {
		t.Error("expected a warning for the bad statement")
	}
}
