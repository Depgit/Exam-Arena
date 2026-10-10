package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"strings"
	"time"
	"unicode"

	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/platform/mailer"
	"github.com/exam-arena/internal/platform/passwords"
	"github.com/exam-arena/internal/platform/tokens"
	"github.com/exam-arena/internal/users"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	userRepo  *users.Store
	jwtSecret string
	jwtExpiry int

	google          *googleVerifier
	codes           *codeStore
	mail            mailer.Mailer
	requireVerified bool
}

// Options are the optional sign-in features.
type Options struct {
	GoogleClientID string        // "" = Google sign-in off
	Mailer         mailer.Mailer // sends verification codes
	// RequireVerifiedEmail: players must verify their email before they
	// can play or chat (demo accounts are blocked anyway).
	RequireVerifiedEmail bool
}

func NewService(userRepo *users.Store, db *pgxpool.Pool, jwtSecret string, jwtExpiry int, opts Options) *Service {
	return &Service{
		userRepo:        userRepo,
		jwtSecret:       jwtSecret,
		jwtExpiry:       jwtExpiry,
		google:          newGoogleVerifier(opts.GoogleClientID),
		codes:           &codeStore{db: db},
		mail:            opts.Mailer,
		requireVerified: opts.RequireVerifiedEmail,
	}
}

// RequireVerifiedEmail reports whether unverified players are kept from playing.
func (s *Service) RequireVerifiedEmail() bool { return s.requireVerified }

// withFlags fills in the computed fields the app needs about a user.
func (s *Service) withFlags(u *models.User) *models.User {
	if u != nil {
		u.NeedsEmailVerification = s.requireVerified && !u.IsGuest && u.Role != "admin" && u.EmailVerifiedAt == nil
	}
	return u
}

// validateUsername applies the username rules shared by every way of signing up.
func validateUsername(name string) error {
	if strings.ContainsFunc(name, unicode.IsSpace) {
		return errors.New("username can't contain spaces")
	}
	if len(name) < 3 || len(name) > 30 {
		return errors.New("username must be between 3 and 30 characters")
	}
	// Reserved prefixes for system accounts (bots, demo guests).
	if lower := strings.ToLower(name); strings.HasPrefix(lower, "bot.") || strings.HasPrefix(lower, "guest_") {
		return errors.New("that username is reserved")
	}
	return nil
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Login    string `json:"login"` // username or email
	Password string `json:"password"`
}

type Response struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (s *Service) Register(ctx context.Context, req RegisterRequest) (*Response, error) {
	// Phone keyboards often add a space after autocomplete. A username saved
	// as "chunnu " can never be found again (lookups trim), so trim here.
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if err := validateUsername(req.Username); err != nil {
		return nil, err
	}
	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}
	if err := validateEmail(req.Email); err != nil {
		return nil, err
	}

	existing, _ := s.userRepo.GetByUsername(ctx, req.Username)
	if existing != nil {
		return nil, errors.New("username already taken")
	}

	existing, _ = s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hash, err := passwords.Hash(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user, err := s.userRepo.Create(ctx, req.Username, req.Email, hash)
	if err != nil {
		return nil, err
	}

	token, err := tokens.Generate(user.ID, user.Username, user.Role, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, err
	}

	// Email the first verification code straight away, without making the
	// sign-up wait for the mail service.
	if s.requireVerified {
		go func(id, email string) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			if err := s.sendCode(ctx, id, email); err != nil {
				slog.Warn("sign-up verification email failed", "user", id, "error", err)
			}
		}(user.ID, user.Email)
	}

	return &Response{Token: token, User: s.withFlags(user)}, nil
}

// Guest creates a fresh throwaway demo account and signs it in.
//
// Every visitor gets their own account: the WebSocket hub allows one live
// session per user, so a single shared demo login would make visitors
// disconnect each other. Guests are hidden from rating leaderboards and
// removed after a few days if nobody else depends on them.
func (s *Service) Guest(ctx context.Context) (*Response, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	// Nobody knows this password; the account is reachable only via the token.
	hash, err := passwords.Hash(hex.EncodeToString(secret))
	if err != nil {
		return nil, errors.New("failed to create guest")
	}

	var user *models.User
	for attempt := 0; attempt < 5 && user == nil; attempt++ {
		n, err := rand.Int(rand.Reader, big.NewInt(900000))
		if err != nil {
			return nil, err
		}
		num := n.Int64() + 100000
		username := fmt.Sprintf("guest_%d", num)
		email := fmt.Sprintf("%s.%s@guest.invalid", username, hex.EncodeToString(secret[:4]))
		user, err = s.userRepo.CreateGuest(ctx, username, email, hash, fmt.Sprintf("Guest %d", num))
		if err != nil && !strings.Contains(err.Error(), "duplicate key") {
			return nil, err
		}
	}
	if user == nil {
		return nil, errors.New("could not create a guest account, please try again")
	}

	token, err := tokens.Generate(user.ID, user.Username, user.Role, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, err
	}
	return &Response{Token: token, User: user}, nil
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*Response, error) {
	var user *models.User
	var err error

	req.Login = strings.TrimSpace(req.Login)
	if strings.Contains(req.Login, "@") {
		user, err = s.userRepo.GetByEmail(ctx, req.Login)
	} else {
		user, err = s.userRepo.GetByUsername(ctx, req.Login)
	}
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}

	if !passwords.Check(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	if user.Status != "active" {
		return nil, errors.New("account is not active")
	}

	token, err := tokens.Generate(user.ID, user.Username, user.Role, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, err
	}

	return &Response{Token: token, User: s.withFlags(user)}, nil
}

func (s *Service) GetUser(ctx context.Context, userID string) (*models.User, error) {
	u, err := s.userRepo.GetByID(ctx, userID)
	return s.withFlags(u), err
}
