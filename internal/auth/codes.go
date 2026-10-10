package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Email verification codes: 6 digits, valid codeTTL, at most
// maxCodeAttempts wrong guesses, a new one at most every resendAfter.
const (
	codeTTL         = 15 * time.Minute
	maxCodeAttempts = 5
	resendAfter     = 60 * time.Second
)

var (
	ErrCodeWrong    = errors.New("that code isn't right")
	ErrCodeExpired  = errors.New("that code has expired — send a new one")
	ErrCodeAttempts = errors.New("too many wrong tries — send a new code")
	ErrCodeNone     = errors.New("no code was sent — send one first")
)

// ErrTooSoon is returned when a code was sent less than resendAfter ago.
type ErrTooSoon struct{ Wait time.Duration }

func (e ErrTooSoon) Error() string {
	return fmt.Sprintf("a code was just sent — try again in %d seconds", int(e.Wait.Seconds())+1)
}

// codeStore keeps one pending code per player, as a hash.
type codeStore struct{ db *pgxpool.Pool }

func newCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// hashCode ties the code to the player, so a stored hash is useless for
// anyone else and can't be looked up in a table of all 10^6 codes at once.
func hashCode(userID, code string) string {
	sum := sha256.Sum256([]byte(userID + ":" + code))
	return hex.EncodeToString(sum[:])
}

// Replace stores a fresh code for the player, unless one was sent less
// than resendAfter ago.
func (s *codeStore) Replace(ctx context.Context, userID, email, code string) error {
	var sentAt time.Time
	err := s.db.QueryRow(ctx, `SELECT sent_at FROM email_verification_codes WHERE user_id = $1`, userID).Scan(&sentAt)
	if err == nil {
		if wait := resendAfter - time.Since(sentAt); wait > 0 {
			return ErrTooSoon{Wait: wait}
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	_, err = s.db.Exec(ctx, `
		INSERT INTO email_verification_codes (user_id, code_hash, email, expires_at, attempts, sent_at)
		VALUES ($1, $2, $3, now() + $4 * interval '1 second', 0, now())
		ON CONFLICT (user_id) DO UPDATE SET code_hash = EXCLUDED.code_hash, email = EXCLUDED.email,
		  expires_at = EXCLUDED.expires_at, attempts = 0, sent_at = now()
	`, userID, hashCode(userID, code), email, int(codeTTL.Seconds()))
	return err
}

// Check compares a guess with the pending code and returns the email it
// was sent to. A wrong guess uses up one attempt.
func (s *codeStore) Check(ctx context.Context, userID, code string) (string, error) {
	var hash, email string
	var expires time.Time
	var attempts int
	err := s.db.QueryRow(ctx, `
		SELECT code_hash, email, expires_at, attempts FROM email_verification_codes WHERE user_id = $1
	`, userID).Scan(&hash, &email, &expires, &attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrCodeNone
	}
	if err != nil {
		return "", err
	}
	switch {
	case attempts >= maxCodeAttempts:
		return "", ErrCodeAttempts
	case time.Now().After(expires):
		return "", ErrCodeExpired
	}
	if subtle.ConstantTimeCompare([]byte(hash), []byte(hashCode(userID, code))) != 1 {
		_, _ = s.db.Exec(ctx, `UPDATE email_verification_codes SET attempts = attempts + 1 WHERE user_id = $1`, userID)
		return "", ErrCodeWrong
	}
	return email, nil
}

// Delete removes the player's pending code once it has been used.
func (s *codeStore) Delete(ctx context.Context, userID string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM email_verification_codes WHERE user_id = $1`, userID)
	return err
}
