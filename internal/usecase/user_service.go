package usecase

import (
	"github.com/studio/platform/configs"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/infra/redis"
)

type EmailSender interface {
	SendPasswordReset(to, username, resetURL string) error
	SendEmailVerification(to, username, verifyURL string) error
}

// UserService handles user-related business logic
type UserService struct {
	userRepo    user.Repository
	tokenStore  *redis.TokenStore
	jwtConfig   configs.JWTConfig
	emailSender EmailSender
	frontendURL string
}

type UserServiceOption func(*UserService)

func WithEmailSender(sender EmailSender, frontendURL string) UserServiceOption {
	return func(s *UserService) {
		s.emailSender = sender
		s.frontendURL = frontendURL
	}
}

// NewUserService creates a new UserService.
func NewUserService(userRepo user.Repository, tokenStore *redis.TokenStore, jwtConfig configs.JWTConfig, opts ...UserServiceOption) *UserService {
	svc := &UserService{
		userRepo:   userRepo,
		tokenStore: tokenStore,
		jwtConfig:  jwtConfig,
	}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

// RegisterInput represents registration input
type RegisterInput struct {
	Username string
	Email    string
	Password string
	IP       string
	Device   string
}

// LoginInput represents login input
type LoginInput struct {
	Email    string
	Password string
	IP       string
	Device   string
}

// TokenOutput represents token output
type TokenOutput struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// AuthOutput represents authentication output
type AuthOutput struct {
	User   *user.User   `json:"user"`
	Tokens *TokenOutput `json:"tokens"`
}
