package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/email"
)

type emailRepo struct {
	db *pgxpool.Pool
}

// NewEmailRepository creates a new email repository
func NewEmailRepository(db *pgxpool.Pool) email.Repository {
	return &emailRepo{db: db}
}

const createEmailSQL = `
	INSERT INTO emails (id, to_address, subject, body, email_type, status, retry_count, error, sent_at, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
`

func (r *emailRepo) CreateEmail(e *email.Email) error {
	_, err := r.db.Exec(context.Background(), createEmailSQL,
		e.ID, e.To, e.Subject, e.Body, e.Type, e.Status,
		e.RetryCount, e.Error, e.SentAt, e.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create email: %w", err)
	}
	return nil
}

const updateEmailSQL = `
	UPDATE emails SET status = $2, retry_count = $3, error = $4, sent_at = $5
	WHERE id = $1
`

func (r *emailRepo) UpdateEmail(e *email.Email) error {
	_, err := r.db.Exec(context.Background(), updateEmailSQL,
		e.ID, e.Status, e.RetryCount, e.Error, e.SentAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}
	return nil
}

const getPendingEmailsSQL = `
	SELECT id, to_address, subject, body, email_type, status, retry_count, error, sent_at, created_at
	FROM emails
	WHERE status = 'pending' AND retry_count < 3
	ORDER BY created_at ASC
	LIMIT $1
`

func (r *emailRepo) GetPendingEmails(limit int) ([]*email.Email, error) {
	rows, err := r.db.Query(context.Background(), getPendingEmailsSQL, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending emails: %w", err)
	}
	defer rows.Close()

	emails := make([]*email.Email, 0)
	for rows.Next() {
		var e email.Email
		err := rows.Scan(
			&e.ID, &e.To, &e.Subject, &e.Body, &e.Type, &e.Status,
			&e.RetryCount, &e.Error, &e.SentAt, &e.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan email: %w", err)
		}
		emails = append(emails, &e)
	}

	return emails, nil
}

const createSubscriptionSQL = `
	INSERT INTO email_subscriptions (id, email, is_verified, verification_token, game_releases, updates, promotions, unsubscribe_token, subscribed_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`

func (r *emailRepo) CreateSubscription(sub *email.Subscription) error {
	_, err := r.db.Exec(context.Background(), createSubscriptionSQL,
		sub.ID, sub.Email, sub.IsVerified, sub.VerificationToken,
		sub.GameReleases, sub.Updates, sub.Promotions,
		sub.UnsubscribeToken, sub.SubscribedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create subscription: %w", err)
	}
	return nil
}

const getSubscriptionSQL = `
	SELECT id, email, is_verified, verification_token, game_releases, updates, promotions, unsubscribe_token, subscribed_at, unsubscribed_at
	FROM email_subscriptions
	WHERE email = $1
`

func (r *emailRepo) GetSubscription(emailAddr string) (*email.Subscription, error) {
	var sub email.Subscription
	err := r.db.QueryRow(context.Background(), getSubscriptionSQL, emailAddr).Scan(
		&sub.ID, &sub.Email, &sub.IsVerified, &sub.VerificationToken,
		&sub.GameReleases, &sub.Updates, &sub.Promotions,
		&sub.UnsubscribeToken, &sub.SubscribedAt, &sub.UnsubscribedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get subscription: %w", err)
	}
	return &sub, nil
}

const updateSubscriptionSQL = `
	UPDATE email_subscriptions
	SET is_verified = $2, game_releases = $3, updates = $4, promotions = $5, unsubscribed_at = $6
	WHERE id = $1
`

func (r *emailRepo) UpdateSubscription(sub *email.Subscription) error {
	_, err := r.db.Exec(context.Background(), updateSubscriptionSQL,
		sub.ID, sub.IsVerified, sub.GameReleases, sub.Updates,
		sub.Promotions, sub.UnsubscribedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update subscription: %w", err)
	}
	return nil
}

const getActiveSubscriptionsSQL = `
	SELECT id, email, is_verified, verification_token, game_releases, updates, promotions, unsubscribe_token, subscribed_at, unsubscribed_at
	FROM email_subscriptions
	WHERE is_verified = TRUE AND unsubscribed_at IS NULL
`

func (r *emailRepo) GetActiveSubscriptions() ([]*email.Subscription, error) {
	rows, err := r.db.Query(context.Background(), getActiveSubscriptionsSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscriptions: %w", err)
	}
	defer rows.Close()

	subs := make([]*email.Subscription, 0)
	for rows.Next() {
		var sub email.Subscription
		err := rows.Scan(
			&sub.ID, &sub.Email, &sub.IsVerified, &sub.VerificationToken,
			&sub.GameReleases, &sub.Updates, &sub.Promotions,
			&sub.UnsubscribeToken, &sub.SubscribedAt, &sub.UnsubscribedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subscription: %w", err)
		}
		subs = append(subs, &sub)
	}

	return subs, nil
}
