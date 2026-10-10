package chat

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store saves chat messages in Postgres (chat_messages).
type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// Save stores a message and fills in its id and timestamp.
func (r *Store) Save(ctx context.Context, m *Message) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO chat_messages (user_id, body) VALUES ($1, $2)
		RETURNING id, created_at
	`, m.UserID, m.Body).Scan(&m.ID, &m.CreatedAt)
}

// Recent returns the newest n messages, oldest first.
func (r *Store) Recent(ctx context.Context, n int) ([]Message, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, username, display_name, body, created_at FROM (
			SELECT c.id, c.user_id, u.username, u.display_name, c.body, c.created_at
			FROM chat_messages c JOIN users u ON u.id = c.user_id
			ORDER BY c.id DESC LIMIT $1
		) newest ORDER BY id ASC
	`, n)
	if err != nil {
		return nil, fmt.Errorf("recent chat: %w", err)
	}
	defer rows.Close()
	out := []Message{}
	for rows.Next() {
		var m Message
		if err := rows.Scan(&m.ID, &m.UserID, &m.Username, &m.DisplayName, &m.Body, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// KeepNewest deletes all but the newest n messages.
func (r *Store) KeepNewest(ctx context.Context, n int) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM chat_messages
		WHERE id < (SELECT coalesce(min(id), 0) FROM (SELECT id FROM chat_messages ORDER BY id DESC LIMIT $1) keep)
	`, n)
	return err
}
