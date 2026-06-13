package generator_test

import (
	"sort"
	"testing"

	"github.com/enricojoe/dgram/backend/internal/model"
	"github.com/enricojoe/dgram/backend/internal/parser"
	"github.com/enricojoe/dgram/backend/internal/parser/generator"
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

// findEnum returns the enum with the given name, or nil.
func findEnum(s model.Schema, name string) *model.Enum {
	for i := range s.Enums {
		if s.Enums[i].Name == name {
			return &s.Enums[i]
		}
	}
	return nil
}

// sampleSchema builds the schema exercised by the round-trip tests. The enum is
// named "<table>_<column>" and attached to posts.status so it round-trips in
// MySQL too (the MySQL parser hoists inline enums into that exact name).
func sampleSchema() model.Schema {
	return model.Schema{
		Tables: []model.Table{
			{
				Name: "users",
				Columns: []model.Column{
					{Name: "id", Type: "serial", PK: true, NotNull: true, AutoInc: true},
					{Name: "email", Type: "varchar(255)", NotNull: true, Unique: true},
				},
			},
			{
				Name: "posts",
				Columns: []model.Column{
					{Name: "id", Type: "serial", PK: true, NotNull: true, AutoInc: true},
					{Name: "user_id", Type: "int", NotNull: true},
					{Name: "status", Type: "posts_status"},
				},
			},
		},
		Refs: []model.Ref{
			{FromTable: "posts", FromColumn: "user_id", ToTable: "users", ToColumn: "id", OnDelete: "CASCADE"},
		},
		Enums: []model.Enum{
			{Name: "posts_status", Values: []string{"draft", "published"}},
		},
	}
}

func TestGenerateUnknownDialect(t *testing.T) {
	if _, err := generator.Generate(model.Dialect("oracle"), model.Schema{}); err == nil {
		t.Fatal("expected error for unknown dialect")
	}
}

func TestRoundTripPostgres(t *testing.T) {
	assertRoundTrip(t, model.DialectPostgres)
}

func TestRoundTripMySQL(t *testing.T) {
	assertRoundTrip(t, model.DialectMySQL)
}

// assertRoundTrip generates DDL from sampleSchema for the dialect, re-parses it,
// and asserts the structurally significant parts survived.
func assertRoundTrip(t *testing.T, d model.Dialect) {
	t.Helper()

	orig := sampleSchema()
	ddl, err := generator.Generate(d, orig)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}
	t.Logf("%s DDL:\n%s", d, ddl)

	got, warnings, err := parser.Parse(d, ddl)
	if err != nil {
		t.Fatalf("re-parse: %v\nDDL:\n%s", err, ddl)
	}
	if len(warnings) != 0 {
		t.Logf("warnings: %v", warnings)
	}

	// Tables.
	if len(got.Tables) != len(orig.Tables) {
		t.Fatalf("table count: want %d, got %d", len(orig.Tables), len(got.Tables))
	}

	users := findTable(got, "users")
	if users == nil {
		t.Fatal("users table missing after round-trip")
	}
	if id := findColumn(users, "id"); id == nil || !id.PK || !id.NotNull || !id.AutoInc {
		t.Errorf("users.id: want pk+notNull+autoInc, got %+v", id)
	}
	if email := findColumn(users, "email"); email == nil || !email.Unique || !email.NotNull {
		t.Errorf("users.email: want unique+notNull, got %+v", email)
	}

	posts := findTable(got, "posts")
	if posts == nil {
		t.Fatal("posts table missing after round-trip")
	}
	if pid := findColumn(posts, "id"); pid == nil || !pid.PK || !pid.AutoInc {
		t.Errorf("posts.id: want pk+autoInc, got %+v", pid)
	}
	if uid := findColumn(posts, "user_id"); uid == nil || !uid.NotNull {
		t.Errorf("posts.user_id: want notNull, got %+v", uid)
	}

	// Foreign key.
	ref := findRef(got, "posts", "user_id")
	if ref == nil {
		t.Fatal("posts.user_id FK missing after round-trip")
	}
	if ref.ToTable != "users" || ref.ToColumn != "id" {
		t.Errorf("FK target: want users.id, got %s.%s", ref.ToTable, ref.ToColumn)
	}
	if ref.OnDelete != "CASCADE" {
		t.Errorf("FK onDelete: want CASCADE, got %q", ref.OnDelete)
	}

	// Enum names + values survive (column linkage is a documented non-goal).
	enum := findEnum(got, "posts_status")
	if enum == nil {
		t.Fatalf("enum posts_status missing after round-trip; got %+v", got.Enums)
	}
	vals := append([]string(nil), enum.Values...)
	sort.Strings(vals)
	if len(vals) != 2 || vals[0] != "draft" || vals[1] != "published" {
		t.Errorf("enum values: want [draft published], got %v", enum.Values)
	}
}
