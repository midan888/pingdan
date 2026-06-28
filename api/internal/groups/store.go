package groups

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Group struct {
	ID        string    `json:"id"`
	UserID    string    `json:"-"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type Store struct{ Pool *pgxpool.Pool }

func (s *Store) List(ctx context.Context, userID string) ([]Group, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT id, user_id, name, created_at
		FROM groups WHERE user_id = $1 ORDER BY name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Group{}
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.UserID, &g.Name, &g.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *Store) Create(ctx context.Context, g *Group) error {
	return s.Pool.QueryRow(ctx, `
		INSERT INTO groups (user_id, name) VALUES ($1, $2)
		RETURNING id, created_at
	`, g.UserID, g.Name).Scan(&g.ID, &g.CreatedAt)
}

func (s *Store) Update(ctx context.Context, userID, id, name string) error {
	_, err := s.Pool.Exec(ctx, `
		UPDATE groups SET name=$1 WHERE id=$2 AND user_id=$3
	`, name, id, userID)
	return err
}

// Delete removes a group. Endpoints in it are left ungrouped (group_id set NULL by FK).
func (s *Store) Delete(ctx context.Context, userID, id string) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM groups WHERE id=$1 AND user_id=$2`, id, userID)
	return err
}

func (s *Store) GetByID(ctx context.Context, id string) (*Group, error) {
	var g Group
	err := s.Pool.QueryRow(ctx, `
		SELECT id, user_id, name, created_at FROM groups WHERE id=$1
	`, id).Scan(&g.ID, &g.UserID, &g.Name, &g.CreatedAt)
	if err == pgx.ErrNoRows {
		return nil, nil
	}
	return &g, err
}
