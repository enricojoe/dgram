package service

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"

	"github.com/enricojoe/dgram/backend/internal/model"
	"github.com/enricojoe/dgram/backend/internal/repository"
)

// ErrDiagramNotFound is returned when a diagram does not exist or is not owned
// by the requesting user (the two are deliberately indistinguishable).
var ErrDiagramNotFound = repository.ErrNotFound

// DiagramService handles diagram CRUD with ownership enforced by passing the
// authenticated user id down to the repository.
type DiagramService struct {
	diagrams *repository.DiagramRepository
}

// NewDiagramService constructs a DiagramService.
func NewDiagramService(diagrams *repository.DiagramRepository) *DiagramService {
	return &DiagramService{diagrams: diagrams}
}

// emptyLayout is the default jsonb value when a caller omits layout.
var emptyLayout = json.RawMessage(`{}`)

// Create stores a new diagram for the user. A nil/empty layout defaults to {}.
func (s *DiagramService) Create(userID int64, name, dialect, ddl string, layout json.RawMessage) (model.Diagram, error) {
	if len(layout) == 0 {
		layout = emptyLayout
	}
	return s.diagrams.Create(userID, name, dialect, ddl, layout)
}

// Get returns a single diagram owned by the user.
func (s *DiagramService) Get(id, userID int64) (model.Diagram, error) {
	return s.diagrams.GetByIDForUser(id, userID)
}

// List returns the user's diagrams in list form.
func (s *DiagramService) List(userID int64) ([]model.DiagramListItem, error) {
	return s.diagrams.ListByUser(userID)
}

// Update applies a partial update to a diagram owned by the user. Any nil field
// is left unchanged.
func (s *DiagramService) Update(id, userID int64, name, dialect, ddl *string, layout json.RawMessage) (model.Diagram, error) {
	return s.diagrams.UpdateForUser(id, userID, name, dialect, ddl, layout)
}

// Delete removes a diagram owned by the user.
func (s *DiagramService) Delete(id, userID int64) error {
	return s.diagrams.DeleteForUser(id, userID)
}

// EnableShare makes a diagram owned by the user public and returns its share
// token. A fresh URL-safe token is generated only when the diagram has none
// yet; an already-shared diagram keeps and returns its existing token.
func (s *DiagramService) EnableShare(id, userID int64) (string, error) {
	token, err := newShareToken()
	if err != nil {
		return "", err
	}
	return s.diagrams.SetShare(userID, id, token, true)
}

// DisableShare turns off public access for a diagram owned by the user while
// preserving its share token so re-sharing reuses the same link.
func (s *DiagramService) DisableShare(id, userID int64) error {
	return s.diagrams.SetPublic(userID, id, false)
}

// GetShared returns the public, read-only view of a diagram by share token.
// Returns ErrDiagramNotFound for unknown or non-public tokens.
func (s *DiagramService) GetShared(shareID string) (model.PublicDiagram, error) {
	d, err := s.diagrams.GetByShareID(shareID)
	if err != nil {
		return model.PublicDiagram{}, err
	}
	return model.PublicDiagram{
		Name:    d.Name,
		Dialect: d.Dialect,
		DDL:     d.DDL,
		Layout:  d.Layout,
	}, nil
}

// newShareToken returns a URL-safe random token (~22 chars) from 16 bytes of
// cryptographically secure randomness.
func newShareToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
