package email

import (
	"time"

	"github.com/google/uuid"
)

// EmailType represents the type of email
type EmailType string

const (
	EmailTypeWelcome          EmailType = "welcome"
	EmailTypeVerification     EmailType = "verification"
	EmailTypePasswordReset    EmailType = "password_reset"
	EmailTypePaymentSuccess   EmailType = "payment_success"
	EmailTypeGameRelease      EmailType = "game_release"
	EmailTypeAchievement      EmailType = "achievement"
	EmailTypeSecurityAlert    EmailType = "security_alert"
	EmailTypeNewsletter       EmailType = "newsletter"
)

// EmailStatus represents the status of an email
type EmailStatus string

const (
	EmailStatusPending EmailStatus = "pending"
	EmailStatusSent    EmailStatus = "sent"
	EmailStatusFailed  EmailStatus = "failed"
)

// Email represents an email message
type Email struct {
	ID          uuid.UUID   `json:"id"`
	To          string      `json:"to"`
	Subject     string      `json:"subject"`
	Body        string      `json:"body"`
	Type        EmailType   `json:"type"`
	Status      EmailStatus `json:"status"`
	RetryCount  int         `json:"retry_count"`
	Error       *string     `json:"error,omitempty"`
	SentAt      *time.Time  `json:"sent_at,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

// Subscription represents an email subscription
type Subscription struct {
	ID              uuid.UUID  `json:"id"`
	Email           string     `json:"email"`
	IsVerified      bool       `json:"is_verified"`
	VerificationToken string   `json:"verification_token"`
	GameReleases    bool       `json:"game_releases"`
	Updates         bool       `json:"updates"`
	Promotions      bool       `json:"promotions"`
	UnsubscribeToken string    `json:"unsubscribe_token"`
	SubscribedAt    time.Time  `json:"subscribed_at"`
	UnsubscribedAt  *time.Time `json:"unsubscribed_at,omitempty"`
}

// Repository defines the interface for email storage
type Repository interface {
	// CreateEmail creates a new email
	CreateEmail(email *Email) error

	// UpdateEmail updates an email
	UpdateEmail(email *Email) error

	// GetPendingEmails retrieves pending emails
	GetPendingEmails(limit int) ([]*Email, error)

	// CreateSubscription creates a new subscription
	CreateSubscription(sub *Subscription) error

	// GetSubscription retrieves a subscription by email
	GetSubscription(email string) (*Subscription, error)

	// UpdateSubscription updates a subscription
	UpdateSubscription(sub *Subscription) error

	// GetActiveSubscriptions retrieves active subscriptions
	GetActiveSubscriptions() ([]*Subscription, error)
}
