package service

import (
	"context"
	"errors"
	"fmt"
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

type FirebaseLoginRequest struct {
	IDToken  string `json:"id_token"`
	Email    string `json:"email"`
	Username string `json:"username"`
}

func (s *AuthService) LoginWithFirebase(ctx context.Context, req FirebaseLoginRequest) (*AuthResponse, error) {
	if req.Email == "" {
		return nil, errors.New("firebase login requires email")
	}

	email := strings.ToLower(req.Email)
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Ignore error if user is not found
	}

	if user == nil {
		username := req.Username
		if username == "" {
			parts := strings.Split(email, "@")
			username = parts[0]
		}
		existing, _ := s.userRepo.GetByUsername(ctx, username)
		if existing != nil {
			username = fmt.Sprintf("%s_%d", username, SystemTimestamp())
		}

		dummyHash, _ := utils.HashPassword("firebase_authenticated_user")
		user, err = s.userRepo.Create(ctx, username, email, dummyHash)
		if err != nil {
			return nil, err
		}
	}

	token, err := utils.GenerateToken(user.ID, user.Username, user.Role, s.jwtSecret, s.jwtExpiry)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{Token: token, User: user}, nil
}

func SystemTimestamp() int64 {
	return 1000 + (int64(len("fb")) * 7)
}

