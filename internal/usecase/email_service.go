package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/studio/platform/internal/domain/email"
	pkgemail "github.com/studio/platform/internal/pkg/email"
	"github.com/studio/platform/internal/pkg/apperr"
)

// EmailService handles email operations
type EmailService struct {
	emailRepo email.Repository
	sender    *pkgemail.Sender
	baseURL   string
}

// NewEmailService creates a new email service
func NewEmailService(emailRepo email.Repository, sender *pkgemail.Sender, baseURL string) *EmailService {
	return &EmailService{
		emailRepo: emailRepo,
		sender:    sender,
		baseURL:   baseURL,
	}
}

// SendWelcomeEmail sends a welcome email to a new user
func (s *EmailService) SendWelcomeEmail(ctx context.Context, to, username string) error {
	body, err := pkgemail.RenderTemplate("welcome", pkgemail.WelcomeEmailData{
		Username: username,
		SiteURL:  s.baseURL,
	})
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.sender.SendEmail(to, "Welcome to Studio Platform!", body, email.EmailTypeWelcome)
}

// SendVerificationEmail sends an email verification email
func (s *EmailService) SendVerificationEmail(ctx context.Context, to, username, token string) error {
	verificationURL := fmt.Sprintf("%s/verify-email?token=%s", s.baseURL, token)

	body, err := pkgemail.RenderTemplate("verification", pkgemail.VerificationEmailData{
		Username:        username,
		VerificationURL: verificationURL,
	})
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.sender.SendEmail(to, "Verify Your Email", body, email.EmailTypeVerification)
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(ctx context.Context, to, username, token string) error {
	resetURL := fmt.Sprintf("%s/reset-password?token=%s", s.baseURL, token)

	body, err := pkgemail.RenderTemplate("password_reset", pkgemail.PasswordResetEmailData{
		Username: username,
		ResetURL: resetURL,
	})
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.sender.SendEmail(to, "Reset Your Password", body, email.EmailTypePasswordReset)
}

// SendGameReleaseEmail sends a game release notification email
func (s *EmailService) SendGameReleaseEmail(ctx context.Context, to, gameTitle, gameID, description string) error {
	gameURL := fmt.Sprintf("%s/games/%s", s.baseURL, gameID)

	body, err := pkgemail.RenderTemplate("game_release", pkgemail.GameReleaseEmailData{
		GameTitle:   gameTitle,
		GameURL:     gameURL,
		Description: description,
	})
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	return s.sender.SendEmail(to, fmt.Sprintf("New Game Released: %s", gameTitle), body, email.EmailTypeGameRelease)
}

// Subscribe creates a new email subscription
func (s *EmailService) Subscribe(ctx context.Context, emailAddr string) error {
	// Check if already subscribed
	existing, err := s.emailRepo.GetSubscription(emailAddr)
	if err == nil && existing != nil {
		if existing.UnsubscribedAt == nil {
			return apperr.New(apperr.CodeInvalidParam, "该邮箱已订阅")
		}
		// Resubscribe
		existing.UnsubscribedAt = nil
		existing.IsVerified = false
		existing.VerificationToken = pkgemail.GenerateToken()
		return s.emailRepo.UpdateSubscription(existing)
	}

	sub := &email.Subscription{
		ID:                uuid.New(),
		Email:             emailAddr,
		IsVerified:        false,
		VerificationToken: pkgemail.GenerateToken(),
		GameReleases:      true,
		Updates:           true,
		Promotions:        false,
		UnsubscribeToken:  pkgemail.GenerateToken(),
		SubscribedAt:      time.Now(),
	}

	if err := s.emailRepo.CreateSubscription(sub); err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "创建订阅失败", err)
	}

	// Send verification email
	verificationURL := fmt.Sprintf("%s/verify-subscription?token=%s", s.baseURL, sub.VerificationToken)
	body := fmt.Sprintf(`
		<h1>Confirm Your Subscription</h1>
		<p>Thank you for subscribing to Studio Platform updates!</p>
		<p>Please confirm your subscription by clicking the link below:</p>
		<p><a href="%s">Confirm Subscription</a></p>
	`, verificationURL)

	return s.sender.SendEmail(emailAddr, "Confirm Your Subscription", body, email.EmailTypeVerification)
}

// VerifySubscription verifies an email subscription
func (s *EmailService) VerifySubscription(ctx context.Context, token string) error {
	subs, err := s.emailRepo.GetActiveSubscriptions()
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "查询订阅失败", err)
	}

	for _, sub := range subs {
		if sub.VerificationToken == token {
			sub.IsVerified = true
			return s.emailRepo.UpdateSubscription(sub)
		}
	}

	return apperr.New(apperr.CodeNotFound, "无效的验证令牌")
}

// Unsubscribe unsubscribes an email
func (s *EmailService) Unsubscribe(ctx context.Context, token string) error {
	subs, err := s.emailRepo.GetActiveSubscriptions()
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "查询订阅失败", err)
	}

	for _, sub := range subs {
		if sub.UnsubscribeToken == token {
			now := time.Now()
			sub.UnsubscribedAt = &now
			return s.emailRepo.UpdateSubscription(sub)
		}
	}

	return apperr.New(apperr.CodeNotFound, "无效的退订令牌")
}

// BroadcastGameRelease broadcasts a game release to all subscribers
func (s *EmailService) BroadcastGameRelease(ctx context.Context, gameTitle, gameID, description string) error {
	subs, err := s.emailRepo.GetActiveSubscriptions()
	if err != nil {
		return apperr.Wrap(apperr.CodeInternalError, "查询订阅失败", err)
	}

	for _, sub := range subs {
		if sub.GameReleases {
			sub := sub // capture loop variable
			go func() {
				if err := s.SendGameReleaseEmail(ctx, sub.Email, gameTitle, gameID, description); err != nil {
					_ = err // best-effort notification; errors are acceptable in goroutines
				}
			}()
		}
	}

	return nil
}
