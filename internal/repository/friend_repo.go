package repository

import (
	"context"
	"fmt"

	"github.com/exam-arena/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type FriendRepo struct {
	db *pgxpool.Pool
}

func NewFriendRepo(db *pgxpool.Pool) *FriendRepo {
	return &FriendRepo{db: db}
}

const friendshipColumns = `id, requester_id, addressee_id, status, created_at, responded_at`

func scanFriendship(row pgx.Row) (*models.Friendship, error) {
	f := &models.Friendship{}
	err := row.Scan(&f.ID, &f.RequesterID, &f.AddresseeID, &f.Status, &f.CreatedAt, &f.RespondedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return f, nil
}

func (r *FriendRepo) GetByID(ctx context.Context, id string) (*models.Friendship, error) {
	f, err := scanFriendship(r.db.QueryRow(ctx, `
		SELECT `+friendshipColumns+` FROM friendships WHERE id = $1
	`, id))
	if err != nil {
		return nil, fmt.Errorf("get friendship: %w", err)
	}
	return f, nil
}

// GetBetween returns the friendship row for the pair in either direction.
func (r *FriendRepo) GetBetween(ctx context.Context, userA, userB string) (*models.Friendship, error) {
	f, err := scanFriendship(r.db.QueryRow(ctx, `
		SELECT `+friendshipColumns+` FROM friendships
		WHERE (requester_id = $1 AND addressee_id = $2)
		   OR (requester_id = $2 AND addressee_id = $1)
	`, userA, userB))
	if err != nil {
		return nil, fmt.Errorf("get friendship between users: %w", err)
	}
	return f, nil
}

func (r *FriendRepo) Create(ctx context.Context, requesterID, addresseeID string) (*models.Friendship, error) {
	f, err := scanFriendship(r.db.QueryRow(ctx, `
		INSERT INTO friendships (requester_id, addressee_id)
		VALUES ($1, $2)
		RETURNING `+friendshipColumns,
		requesterID, addresseeID))
	if err != nil {
		return nil, fmt.Errorf("create friendship: %w", err)
	}
	return f, nil
}

// Reopen turns a declined row back into a pending request from requesterID.
func (r *FriendRepo) Reopen(ctx context.Context, id, requesterID, addresseeID string) (*models.Friendship, error) {
	f, err := scanFriendship(r.db.QueryRow(ctx, `
		UPDATE friendships
		SET requester_id = $2, addressee_id = $3, status = 'pending',
		    created_at = now(), responded_at = NULL
		WHERE id = $1 AND status = 'declined'
		RETURNING `+friendshipColumns,
		id, requesterID, addresseeID))
	if err != nil {
		return nil, fmt.Errorf("reopen friendship: %w", err)
	}
	return f, nil
}

// Respond moves a pending request to accepted or declined. It returns nil
// when the row is no longer pending (already answered or withdrawn).
func (r *FriendRepo) Respond(ctx context.Context, id, status string) (*models.Friendship, error) {
	f, err := scanFriendship(r.db.QueryRow(ctx, `
		UPDATE friendships
		SET status = $2, responded_at = now()
		WHERE id = $1 AND status = 'pending'
		RETURNING `+friendshipColumns,
		id, status))
	if err != nil {
		return nil, fmt.Errorf("respond to friendship: %w", err)
	}
	return f, nil
}

func (r *FriendRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM friendships WHERE id = $1`, id)
	return err
}

// ListForUser returns every pending or accepted friendship involving userID,
// described from userID's side. Online is left for the caller to fill in.
func (r *FriendRepo) ListForUser(ctx context.Context, userID string) ([]models.FriendEntry, error) {
	rows, err := r.db.Query(ctx, `
		SELECT f.id, u.id, u.username, u.display_name, u.avatar_url, f.status,
		       CASE WHEN f.requester_id = $1 THEN 'outgoing' ELSE 'incoming' END,
		       f.created_at, f.responded_at
		FROM friendships f
		JOIN users u ON u.id = CASE WHEN f.requester_id = $1 THEN f.addressee_id ELSE f.requester_id END
		WHERE (f.requester_id = $1 OR f.addressee_id = $1)
		  AND f.status IN ('pending', 'accepted')
		  AND u.deleted_at IS NULL
		ORDER BY u.username
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list friendships: %w", err)
	}
	defer rows.Close()

	var entries []models.FriendEntry
	for rows.Next() {
		var e models.FriendEntry
		if err := rows.Scan(
			&e.FriendshipID, &e.UserID, &e.Username, &e.DisplayName, &e.AvatarURL,
			&e.Status, &e.Direction, &e.CreatedAt, &e.RespondedAt,
		); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
