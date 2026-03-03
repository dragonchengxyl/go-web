package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/studio/platform/internal/domain/user"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

const createUserSQL = `
	INSERT INTO users (id, username, email, password_hash, role, status, created_at, updated_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
`

func (r *UserRepository) Create(ctx context.Context, u *user.User) error {
	_, err := r.pool.Exec(ctx, createUserSQL,
		u.ID, u.Username, u.Email, u.PasswordHash, u.Role, u.Status, u.CreatedAt, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

const getUserByIDSQL = `
	SELECT id, username, email, password_hash, avatar_key, bio, location, role, status,
	       email_verified_at, last_login_at, last_login_ip, created_at, updated_at
	FROM users WHERE id = $1
`

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var u user.User
	err := r.pool.QueryRow(ctx, getUserByIDSQL, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.AvatarKey, &u.Bio, &u.Location,
		&u.Role, &u.Status, &u.EmailVerifiedAt, &u.LastLoginAt, &u.LastLoginIP,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &u, nil
}

const getUserByEmailSQL = `
	SELECT id, username, email, password_hash, avatar_key, bio, location, role, status,
	       email_verified_at, last_login_at, last_login_ip, created_at, updated_at
	FROM users WHERE email = $1
`

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.pool.QueryRow(ctx, getUserByEmailSQL, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.AvatarKey, &u.Bio, &u.Location,
		&u.Role, &u.Status, &u.EmailVerifiedAt, &u.LastLoginAt, &u.LastLoginIP,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}
	return &u, nil
}

const getUserByUsernameSQL = `
	SELECT id, username, email, password_hash, avatar_key, bio, location, role, status,
	       email_verified_at, last_login_at, last_login_ip, created_at, updated_at
	FROM users WHERE username = $1
`

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	var u user.User
	err := r.pool.QueryRow(ctx, getUserByUsernameSQL, username).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.AvatarKey, &u.Bio, &u.Location,
		&u.Role, &u.Status, &u.EmailVerifiedAt, &u.LastLoginAt, &u.LastLoginIP,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &u, nil
}

const updateUserSQL = `
	UPDATE users
	SET username = $2, email = $3, avatar_key = $4, bio = $5, location = $6,
	    role = $7, status = $8, updated_at = $9
	WHERE id = $1
`

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	_, err := r.pool.Exec(ctx, updateUserSQL,
		u.ID, u.Username, u.Email, u.AvatarKey, u.Bio, u.Location,
		u.Role, u.Status, u.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

const updateLastLoginSQL = `
	UPDATE users
	SET last_login_at = NOW(), last_login_ip = $2
	WHERE id = $1
`

func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error {
	_, err := r.pool.Exec(ctx, updateLastLoginSQL, id, ip)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}
	return nil
}

const existsByEmailSQL = `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, existsByEmailSQL, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check email existence: %w", err)
	}
	return exists, nil
}

const existsByUsernameSQL = `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`

func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, existsByUsernameSQL, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check username existence: %w", err)
	}
	return exists, nil
}
