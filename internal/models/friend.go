package models

import "time"

type Friendship struct {
	ID          string     `json:"id"`
	RequesterID string     `json:"requester_id"`
	AddresseeID string     `json:"addressee_id"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	RespondedAt *time.Time `json:"responded_at"`
}

// FriendEntry is one row of a user's friends list, seen from that user's side:
// the fields describe the *other* user in the friendship.
type FriendEntry struct {
	FriendshipID string     `json:"friendship_id"`
	UserID       string     `json:"user_id"`
	Username     string     `json:"username"`
	DisplayName  *string    `json:"display_name"`
	AvatarURL    *string    `json:"avatar_url"`
	Status       string     `json:"status"`    // pending | accepted
	Direction    string     `json:"direction"` // incoming | outgoing (who sent the request)
	Online       bool       `json:"online"`
	CreatedAt    time.Time  `json:"created_at"`
	RespondedAt  *time.Time `json:"responded_at"`
}

type FriendsList struct {
	Friends  []FriendEntry `json:"friends"`
	Incoming []FriendEntry `json:"incoming"`
	Outgoing []FriendEntry `json:"outgoing"`
}
