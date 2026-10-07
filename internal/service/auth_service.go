package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/exam-arena/internal/models"
	"github.com/exam-arena/internal/repository"
	"github.com/exam-arena/internal/utils"
)

type AuthService struct {
	userRepo  *repository.UserRepo
	jwtSecret string
	jwtExpiry int
}

func NewAuthService(userRepo *repository.UserRepo, jwtSecret string, jwtExpiry int) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		jwtExpiry: jwtExpiry,
	}
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

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	if len(req.Username) < 3 || len(req.Username) > 30 {
		return nil, errors.New("username must be between 3 and 30 characters")
	}
	if len(req.Password) < 8 {
		return nil, errors.New("password must be at least 8 characters")
	}
	if !strings.Contains(req.Email, "@") {
		return nil, errors.New("invalid email address")
	}

	existing, _ := s.userRepo.GetByUsername(ctx, req.Username)
	if existing != nil {
		return nil, errors.New("username already taken")
	}

	existing, _ = s.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user, err := s.userRepo.Create(ctx, req.Username, strings.ToLower(req.Email), hash)
	if err != nil {
		return nil, err
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

// Guest creates a fresh throwaway demo account and signs it in.
//
// Every visitor gets their own account: the WebSocket hub allows one live
// session per user, so a single shared demo login would make visitors
// disconnect each other. Guests are hidden from rating leaderboards and
// removed after a few days if nobody else depends on them.
func (s *AuthService) Guest(ctx context.Context) (*AuthResponse, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, err
	}
	// Nobody knows this password; the account is reachable only via the token.
	hash, err := utils.HashPassword(hex.EncodeToString(secret))
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

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, err
	}
	return &AuthResponse{Token: token, User: user}, nil
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	var user *models.User
	var err error

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

	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	if user.Status != "active" {
		return nil, errors.New("account is not active")
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

func (s *AuthService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
