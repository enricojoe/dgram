package model

import (
	"encoding/json"
	"time"
)

// Diagram is a saved schema diagram owned by a user. Layout is an arbitrary
// JSON object (positions keyed by table name) stored verbatim as Postgres jsonb.
type Diagram struct {
	ID        int64           `json:"id" db:"id"`
	UserID    int64           `json:"-" db:"user_id"`
	Name      string          `json:"name" db:"name"`
	Dialect   string          `json:"dialect" db:"dialect"`
	DDL       string          `json:"ddl" db:"ddl_source"`
	Layout    json.RawMessage `json:"layout" db:"layout"`
	ShareID   *string         `json:"shareId,omitempty" db:"share_id"`
	IsPublic  bool            `json:"isPublic" db:"is_public"`
	CreatedAt time.Time       `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time       `json:"updatedAt" db:"updated_at"`
}

// PublicDiagram is the read-only projection served to anonymous visitors via a
// share link. It deliberately omits id and any owner information.
type PublicDiagram struct {
	Name    string          `json:"name" db:"name"`
	Dialect string          `json:"dialect" db:"dialect"`
	DDL     string          `json:"ddl" db:"ddl_source"`
	Layout  json.RawMessage `json:"layout" db:"layout"`
}

// DiagramListItem is the lightweight list projection, omitting the heavy
// ddl/layout payload.
type DiagramListItem struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Dialect   string    `json:"dialect" db:"dialect"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}
