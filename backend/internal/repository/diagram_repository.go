package repository

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/enricojoe/dgram/backend/internal/model"
)

// DiagramRepository persists and retrieves diagrams. Every query is scoped by
// user_id so a user can never read or mutate another user's rows.
type DiagramRepository struct {
	db *sqlx.DB
}

// NewDiagramRepository constructs a DiagramRepository.
func NewDiagramRepository(db *sqlx.DB) *DiagramRepository {
	return &DiagramRepository{db: db}
}

// Create inserts a new diagram owned by userID and returns the stored row.
func (r *DiagramRepository) Create(userID int64, name, dialect, ddl string, layout json.RawMessage) (model.Diagram, error) {
	var d model.Diagram
	const q = `
		INSERT INTO diagrams (user_id, name, dialect, ddl_source, layout)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, name, dialect, ddl_source, layout, share_id, is_public, created_at, updated_at`
	err := r.db.Get(&d, q, userID, name, dialect, ddl, []byte(layout))
	return d, err
}

// GetByIDForUser fetches a single diagram owned by userID. Returns ErrNotFound
// when the diagram does not exist or belongs to another user.
func (r *DiagramRepository) GetByIDForUser(id, userID int64) (model.Diagram, error) {
	var d model.Diagram
	const q = `
		SELECT id, user_id, name, dialect, ddl_source, layout, share_id, is_public, created_at, updated_at
		FROM diagrams
		WHERE id = $1 AND user_id = $2`
	if err := r.db.Get(&d, q, id, userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Diagram{}, ErrNotFound
		}
		return model.Diagram{}, err
	}
	return d, nil
}

// ListByUser returns the lightweight list view of all diagrams owned by userID,
// most recently updated first.
func (r *DiagramRepository) ListByUser(userID int64) ([]model.DiagramListItem, error) {
	items := []model.DiagramListItem{}
	const q = `
		SELECT id, name, dialect, created_at, updated_at
		FROM diagrams
		WHERE user_id = $1
		ORDER BY updated_at DESC`
	if err := r.db.Select(&items, q, userID); err != nil {
		return nil, err
	}
	return items, nil
}

// UpdateForUser applies the given fields to a diagram owned by userID. Any nil
// pointer leaves the corresponding column unchanged. Returns ErrNotFound when
// the diagram does not exist or belongs to another user.
func (r *DiagramRepository) UpdateForUser(id, userID int64, name, dialect, ddl *string, layout json.RawMessage) (model.Diagram, error) {
	var d model.Diagram
	const q = `
		UPDATE diagrams SET
			name       = COALESCE($3, name),
			dialect    = COALESCE($4, dialect),
			ddl_source = COALESCE($5, ddl_source),
			layout     = COALESCE($6, layout),
			updated_at = now()
		WHERE id = $1 AND user_id = $2
		RETURNING id, user_id, name, dialect, ddl_source, layout, share_id, is_public, created_at, updated_at`

	var layoutArg any
	if layout != nil {
		layoutArg = []byte(layout)
	}
	if err := r.db.Get(&d, q, id, userID, name, dialect, ddl, layoutArg); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Diagram{}, ErrNotFound
		}
		return model.Diagram{}, err
	}
	return d, nil
}

// DeleteForUser removes a diagram owned by userID. Returns ErrNotFound when the
// diagram does not exist or belongs to another user.
func (r *DiagramRepository) DeleteForUser(id, userID int64) error {
	const q = `DELETE FROM diagrams WHERE id = $1 AND user_id = $2`
	res, err := r.db.Exec(q, id, userID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// SetShare enables sharing for a diagram owned by userID. The share_id is only
// assigned when currently null (COALESCE) so an existing token is reused and
// stays stable across re-shares. Returns the effective share_id. Returns
// ErrNotFound when the diagram does not exist or belongs to another user.
func (r *DiagramRepository) SetShare(userID, id int64, shareID string, public bool) (string, error) {
	var effective string
	const q = `
		UPDATE diagrams SET
			share_id  = COALESCE(share_id, $3),
			is_public = $4
		WHERE id = $1 AND user_id = $2
		RETURNING share_id`
	if err := r.db.Get(&effective, q, id, userID, shareID, public); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotFound
		}
		return "", err
	}
	return effective, nil
}

// SetPublic toggles the is_public flag on a diagram owned by userID while
// leaving share_id untouched. Returns ErrNotFound when the diagram does not
// exist or belongs to another user.
func (r *DiagramRepository) SetPublic(userID, id int64, public bool) error {
	const q = `UPDATE diagrams SET is_public = $3 WHERE id = $1 AND user_id = $2`
	res, err := r.db.Exec(q, id, userID, public)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// GetByShareID fetches a publicly shared diagram by its share token. It is NOT
// user-scoped but only returns rows that are currently public, so private or
// unknown share ids are indistinguishable (both yield ErrNotFound).
func (r *DiagramRepository) GetByShareID(shareID string) (model.Diagram, error) {
	var d model.Diagram
	const q = `
		SELECT id, user_id, name, dialect, ddl_source, layout, share_id, is_public, created_at, updated_at
		FROM diagrams
		WHERE share_id = $1 AND is_public = true`
	if err := r.db.Get(&d, q, shareID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Diagram{}, ErrNotFound
		}
		return model.Diagram{}, err
	}
	return d, nil
}
