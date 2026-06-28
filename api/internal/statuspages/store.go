package statuspages

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Page is a status page owned by a user, addressed publicly by Slug.
type Page struct {
	ID          string    `json:"id"`
	UserID      string    `json:"-"`
	Slug        string    `json:"slug"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Item is one endpoint included on a page (owner-facing view).
type Item struct {
	EndpointID  string  `json:"endpointId"`
	DisplayName *string `json:"displayName"`
	Position    int     `json:"position"`
}

type Store struct{ Pool *pgxpool.Pool }

func (s *Store) List(ctx context.Context, userID string) ([]Page, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, user_id, slug, title, description, created_at, updated_at
		FROM status_pages WHERE user_id = $1 ORDER BY created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Page{}
	for rows.Next() {
		var p Page
		if err := rows.Scan(&p.ID, &p.UserID, &p.Slug, &p.Title, &p.Description, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) Create(ctx context.Context, p *Page) error {
	return s.Pool.QueryRow(ctx, `
		INSERT INTO status_pages (user_id, slug, title, description)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`, p.UserID, p.Slug, p.Title, p.Description).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (s *Store) Update(ctx context.Context, userID, id, slug, title, description string) error {
	_, err := s.Pool.Exec(ctx, `
		UPDATE status_pages SET slug=$1, title=$2, description=$3, updated_at=now()
		WHERE id=$4 AND user_id=$5
	`, slug, title, description, id, userID)
	return err
}

func (s *Store) Delete(ctx context.Context, userID, id string) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM status_pages WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

// GetByID returns a page by id, or nil if it does not exist.
func (s *Store) GetByID(ctx context.Context, id string) (*Page, error) {
	var p Page
	err := s.Pool.QueryRow(ctx, `
		SELECT id, user_id, slug, title, description, created_at, updated_at
		FROM status_pages WHERE id=$1
	`, id).Scan(&p.ID, &p.UserID, &p.Slug, &p.Title, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

// GetBySlug returns a page by its public slug, or nil if it does not exist.
func (s *Store) GetBySlug(ctx context.Context, slug string) (*Page, error) {
	var p Page
	err := s.Pool.QueryRow(ctx, `
		SELECT id, user_id, slug, title, description, created_at, updated_at
		FROM status_pages WHERE slug=$1
	`, slug).Scan(&p.ID, &p.UserID, &p.Slug, &p.Title, &p.Description, &p.CreatedAt, &p.UpdatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

// Items returns the endpoint membership for a page, ordered by position.
func (s *Store) Items(ctx context.Context, pageID string) ([]Item, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT endpoint_id, display_name, position
		FROM status_page_endpoints WHERE page_id=$1 ORDER BY position ASC, endpoint_id ASC
	`, pageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Item{}
	for rows.Next() {
		var it Item
		if err := rows.Scan(&it.EndpointID, &it.DisplayName, &it.Position); err != nil {
			return nil, err
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

// SetItems replaces the full endpoint membership for a page in one transaction.
// Each item's endpoint is verified to belong to userID, so a page can never
// include an endpoint the owner does not own.
func (s *Store) SetItems(ctx context.Context, userID, pageID string, items []Item) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM status_page_endpoints WHERE page_id=$1`, pageID); err != nil {
		return err
	}
	for i, it := range items {
		// Insert only if the endpoint is owned by this user; the subselect
		// guards against attaching someone else's endpoint.
		if _, err := tx.Exec(ctx, `
			INSERT INTO status_page_endpoints (page_id, endpoint_id, display_name, position)
			SELECT $1, e.id, $2, $3 FROM endpoints e WHERE e.id=$4 AND e.user_id=$5
		`, pageID, nullIfEmpty(it.DisplayName), i, it.EndpointID, userID); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func nullIfEmpty(s *string) *string {
	if s == nil {
		return nil
	}
	if *s == "" {
		return nil
	}
	return s
}
