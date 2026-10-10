package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode"

	"github.com/exam-arena/internal/platform/passwords"
	"github.com/exam-arena/internal/platform/tokens"
)

// GoogleRequest is what the app sends after the "Sign in with Google"
// button: Google's ID token, plus a username on the second step of a
// first-time sign-up.
type GoogleRequest struct {
	Credential string `json:"credential"`
	Username   string `json:"username,omitempty"`
}

// GoogleResult is either a signed-in session, or — for someone new — a
// request to pick a username first (send the same credential back with it).
type GoogleResult struct {
	*Response
	NeedsUsername     bool   `json:"needs_username,omitempty"`
	SuggestedUsername string `json:"suggested_username,omitempty"`
	Email             string `json:"email,omitempty"`
	Name              string `json:"name,omitempty"`
}

// Google signs a player in with their Google account:
//
//  1. already linked to a player   → signed in
//  2. email matches a player        → Google is linked to that account, signed in
//     (Google has verified the email, so it is the same person)
//  3. new                           → asks for a username, then creates the player
func (s *Service) Google(ctx context.Context, req GoogleRequest) (*GoogleResult, error) {
	id, err := s.google.Verify(ctx, req.Credential)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByGoogleSub(ctx, id.Sub)
	if err != nil {
		return nil, err
	}
	if user == nil {
		existing, err := s.userRepo.GetByEmail(ctx, id.Email)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			if err := s.userRepo.LinkGoogle(ctx, existing.ID, id.Sub); err != nil {
				return nil, err
			}
			if user, err = s.userRepo.GetByID(ctx, existing.ID); err != nil {
				return nil, err
			}
		}
	}

	if user == nil {
		username := strings.TrimSpace(req.Username)
		if username == "" {
			return &GoogleResult{
				NeedsUsername:     true,
				SuggestedUsername: s.suggestUsername(ctx, id),
				Email:             id.Email,
				Name:              id.Name,
			}, nil
		}
		if err := validateUsername(username); err != nil {
			return nil, err
		}
		if taken, _ := s.userRepo.GetByUsername(ctx, username); taken != nil {
			return nil, errors.New("username already taken")
		}
		secret := make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return nil, err
		}
		hash, err := passwords.Hash(hex.EncodeToString(secret))
		if err != nil {
			return nil, errors.New("failed to create account")
		}
		display := strings.TrimSpace(id.Name)
		if display == "" || len(display) > 60 {
			display = username
		}
		user, err = s.userRepo.CreateFromGoogle(ctx, username, id.Email, hash, display, id.Picture, id.Sub)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				return nil, errors.New("username already taken")
			}
			return nil, err
		}
	}

	if user.Status != "active" {
		return nil, errors.New("account is not active")
	}
	token, err := tokens.Generate(user.ID, user.Username, user.Role, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, err
	}
	return &GoogleResult{Response: &Response{Token: token, User: s.withFlags(user)}}, nil
}

// suggestUsername offers a free username based on the email's local part
// ("priya.sharma@…" → "priya.sharma", or "priya.sharma7" when taken).
func (s *Service) suggestUsername(ctx context.Context, id *GoogleIdentity) string {
	local, _, _ := strings.Cut(id.Email, "@")
	base := strings.Map(func(r rune) rune {
		switch {
		case r < 128 && (unicode.IsLetter(r) || unicode.IsDigit(r)), r == '_', r == '.':
			return unicode.ToLower(r)
		}
		return -1
	}, local)
	if len(base) > 24 {
		base = base[:24]
	}
	for len(base) < 3 {
		base += "x"
	}
	if validateUsername(base) != nil {
		base = "player"
	}
	candidate := base
	for i := 0; i < 20; i++ {
		if taken, _ := s.userRepo.GetByUsername(ctx, candidate); taken == nil {
			return candidate
		}
		n := make([]byte, 2)
		_, _ = rand.Read(n)
		candidate = fmt.Sprintf("%s%d", base, (int(n[0])<<8|int(n[1]))%1000)
	}
	return candidate
}
