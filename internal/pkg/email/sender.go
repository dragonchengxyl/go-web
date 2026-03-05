package email

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/smtp"
	"time"

	"github.com/google/uuid"
	domainEmail "github.com/studio/platform/internal/domain/email"
	"go.uber.org/zap"
)

// SMTPConfig represents SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

// Sender handles email sending
type Sender struct {
	config SMTPConfig
	repo   domainEmail.Repository
	logger *zap.Logger
}

// NewSender creates a new email sender
func NewSender(config SMTPConfig, repo domainEmail.Repository, logger *zap.Logger) *Sender {
	return &Sender{
		config: config,
		repo:   repo,
		logger: logger,
	}
}

// SendEmail sends an email
func (s *Sender) SendEmail(to, subject, body string, emailType domainEmail.EmailType) error {
	email := &domainEmail.Email{
		ID:        uuid.New(),
		To:        to,
		Subject:   subject,
		Body:      body,
		Type:      emailType,
		Status:    domainEmail.EmailStatusPending,
		CreatedAt: time.Now(),
	}

	if err := s.repo.CreateEmail(email); err != nil {
		return fmt.Errorf("failed to queue email: %w", err)
	}

	// Send asynchronously
	go s.processPendingEmails()

	return nil
}

// processPendingEmails processes pending emails
func (s *Sender) processPendingEmails() {
	emails, err := s.repo.GetPendingEmails(10)
	if err != nil {
		s.logger.Error("Failed to get pending emails", zap.Error(err))
		return
	}

	for _, email := range emails {
		if err := s.send(email); err != nil {
			s.logger.Error("Failed to send email",
				zap.String("email_id", email.ID.String()),
				zap.String("to", email.To),
				zap.Error(err))

			email.Status = domainEmail.EmailStatusFailed
			email.RetryCount++
			errMsg := err.Error()
			email.Error = &errMsg
		} else {
			email.Status = domainEmail.EmailStatusSent
			now := time.Now()
			email.SentAt = &now
		}

		if err := s.repo.UpdateEmail(email); err != nil {
			s.logger.Error("Failed to update email status", zap.Error(err))
		}
	}
}

// send sends an email via SMTP
func (s *Sender) send(e *domainEmail.Email) error {
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n",
		s.config.From, e.To, e.Subject, e.Body)

	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	return smtp.SendMail(addr, auth, s.config.From, []string{e.To}, []byte(msg))
}

// Template data structures
type WelcomeEmailData struct {
	Username string
	SiteURL  string
}

type VerificationEmailData struct {
	Username        string
	VerificationURL string
}

type PasswordResetEmailData struct {
	Username string
	ResetURL string
}

type GameReleaseEmailData struct {
	GameTitle   string
	GameURL     string
	Description string
}

// RenderTemplate renders an email template
func RenderTemplate(templateName string, data any) (string, error) {
	templates := map[string]string{
		"welcome": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Welcome</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #4a5568;">Welcome to Studio Platform!</h1>
        <p>Hi {{.Username}},</p>
        <p>Thank you for joining our community! We're excited to have you here.</p>
        <p>Explore our latest games and music at <a href="{{.SiteURL}}" style="color: #3182ce;">{{.SiteURL}}</a></p>
        <p>Best regards,<br>The Studio Platform Team</p>
    </div>
</body>
</html>`,
		"verification": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Email Verification</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #4a5568;">Verify Your Email</h1>
        <p>Hi {{.Username}},</p>
        <p>Please verify your email address by clicking the button below:</p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="{{.VerificationURL}}" style="background-color: #3182ce; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Verify Email</a>
        </p>
        <p>If you didn't create an account, you can safely ignore this email.</p>
        <p>Best regards,<br>The Studio Platform Team</p>
    </div>
</body>
</html>`,
		"password_reset": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Password Reset</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #4a5568;">Reset Your Password</h1>
        <p>Hi {{.Username}},</p>
        <p>We received a request to reset your password. Click the button below to create a new password:</p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="{{.ResetURL}}" style="background-color: #3182ce; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">Reset Password</a>
        </p>
        <p>This link will expire in 1 hour.</p>
        <p>If you didn't request a password reset, please ignore this email.</p>
        <p>Best regards,<br>The Studio Platform Team</p>
    </div>
</body>
</html>`,
		"game_release": `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>New Game Release</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #4a5568;">New Game Released!</h1>
        <h2>{{.GameTitle}}</h2>
        <p>{{.Description}}</p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="{{.GameURL}}" style="background-color: #3182ce; color: white; padding: 12px 24px; text-decoration: none; border-radius: 4px; display: inline-block;">View Game</a>
        </p>
        <p>Best regards,<br>The Studio Platform Team</p>
    </div>
</body>
</html>`,
	}

	tmplStr, ok := templates[templateName]
	if !ok {
		return "", fmt.Errorf("template not found: %s", templateName)
	}

	tmpl, err := template.New(templateName).Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// GenerateToken generates a random token
func GenerateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}
