package dialect

import (
	"testing"
)

const mysqlDDL = "" +
	"CREATE TABLE `users` (" +
	"  `id` bigint NOT NULL AUTO_INCREMENT," +
	"  `email` varchar(255) NOT NULL," +
	"  `role` enum('admin','user') DEFAULT 'user'," +
	"  PRIMARY KEY (`id`)," +
	"  UNIQUE KEY `uq_email` (`email`)" +
	") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;" +
	"CREATE TABLE `posts` (" +
	"  `id` bigint NOT NULL AUTO_INCREMENT," +
	"  `user_id` bigint NOT NULL," +
	"  `title` varchar(100) NOT NULL," +
	"  PRIMARY KEY (`id`)," +
	"  CONSTRAINT `fk_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE RESTRICT" +
	") ENGINE=InnoDB;"

func TestParseMySQL(t *testing.T) {
	schema, warnings, err := ParseMySQL(mysqlDDL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(warnings) != 0 {
		t.Logf("warnings: %v", warnings)
	}

	if len(schema.Tables) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(schema.Tables))
	}

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
	if id != nil && id.Type != "bigint" {
		t.Errorf("users.id: want type bigint (width stripped), got %q", id.Type)
	}
	email := findColumn(users, "email")
	if email == nil || !email.Unique || !email.NotNull {
		t.Errorf("users.email: want unique+notNull, got %+v", email)
	}

	// inline enum hoisted to a named type.
	role := findColumn(users, "role")
	if role == nil {
		t.Fatal("users.role not found")
	}
	if role.Type != "users_role" {
		t.Errorf("users.role: want type users_role, got %q", role.Type)
	}
	if len(schema.Enums) != 1 {
		t.Fatalf("expected 1 enum, got %d", len(schema.Enums))
	}
	if got := schema.Enums[0].Values; len(got) != 2 || got[0] != "admin" || got[1] != "user" {
		t.Errorf("enum values: want [admin user], got %v", got)
	}

	// posts FK.
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
	if ref.OnUpdate != "RESTRICT" {
		t.Errorf("FK onUpdate: want RESTRICT, got %q", ref.OnUpdate)
	}
}

func TestParseMySQLDefault(t *testing.T) {
	schema, _, err := ParseMySQL("CREATE TABLE t (id bigint AUTO_INCREMENT PRIMARY KEY, status varchar(20) DEFAULT 'active');")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	status := findColumn(findTable(schema, "t"), "status")
	if status == nil || status.Default == nil {
		t.Fatalf("status default missing: %+v", status)
	}
	if *status.Default != "'active'" {
		t.Errorf("status default: want 'active', got %q", *status.Default)
	}
}
