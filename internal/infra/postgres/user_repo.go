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
	SELECT id, username, email, password_hash, avatar_key, bio, website, location, furry_name, species, role, status,
	       email_verified_at, last_login_at, last_login_ip, created_at, updated_at
	FROM users WHERE id = $1
`

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	var u user.User
	err := r.pool.QueryRow(ctx, getUserByIDSQL, id).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.AvatarKey, &u.Bio, &u.Website, &u.Location,
		&u.FurryName, &u.Species,
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
	SELECT id, username, email, password_hash, avatar_key, bio, website, location, furry_name, species, role, status,
	       email_verified_at, last_login_at, last_login_ip, created_at, updated_at
	FROM users WHERE email = $1
`

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.pool.QueryRow(ctx, getUserByEmailSQL, email).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.AvatarKey, &u.Bio, &u.Website, &u.Location,
		&u.FurryName, &u.Species,
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
	SELECT id, username, email, password_hash, avatar_key, bio, website, location, furry_name, species, role, status,
	       email_verified_at, last_login_at, last_login_ip, created_at, updated_at
	FROM users WHERE username = $1
`

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*user.User, error) {
	var u user.User
	err := r.pool.QueryRow(ctx, getUserByUsernameSQL, username).Scan(
		&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.AvatarKey, &u.Bio, &u.Website, &u.Location,
		&u.FurryName, &u.Species,
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
	SET username = $2, email = $3, avatar_key = $4, bio = $5, website = $6, location = $7,
	    furry_name = $8, species = $9, role = $10, status = $11, updated_at = $12
	WHERE id = $1
`

func (r *UserRepository) Update(ctx context.Context, u *user.User) error {
	_, err := r.pool.Exec(ctx, updateUserSQL,
		u.ID, u.Username, u.Email, u.AvatarKey, u.Bio, u.Website, u.Location,
		u.FurryName, u.Species, u.Role, u.Status, u.UpdatedAt,
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

// List retrieves users with pagination and filters
func (r *UserRepository) List(ctx context.Context, filter user.ListFilter) ([]*user.User, int64, error) {
	// Build query
	query := `
		SELECT id, username, email, password_hash, avatar_key, bio, website, location, furry_name, species, role, status,
		       email_verified_at, last_login_at, last_login_ip, created_at, updated_at
		FROM users
		WHERE 1=1
	`
	countQuery := `SELECT COUNT(*) FROM users WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	// Apply filters
	if filter.Role != nil {
		query += fmt.Sprintf(" AND role = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND role = $%d", argIndex)
		args = append(args, *filter.Role)
		argIndex++
	}

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		countQuery += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		query += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex)
		countQuery += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d)", argIndex, argIndex)
		args = append(args, searchPattern)
		argIndex++
	}

	// Get total count
	var total int64
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	// Apply pagination
	query += " ORDER BY created_at DESC"
	if filter.PageSize > 0 {
		offset := (filter.Page - 1) * filter.PageSize
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
		args = append(args, filter.PageSize, offset)
	}

	// Execute query
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	users := make([]*user.User, 0)
	for rows.Next() {
		var u user.User
		err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.PasswordHash, &u.AvatarKey, &u.Bio, &u.Website, &u.Location,
			&u.FurryName, &u.Species,
			&u.Role, &u.Status, &u.EmailVerifiedAt, &u.LastLoginAt, &u.LastLoginIP,
			&u.CreatedAt, &u.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, &u)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return users, total, nil
}

const deleteUserSQL = `DELETE FROM users WHERE id = $1`

// Delete deletes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.pool.Exec(ctx, deleteUserSQL, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return user.ErrNotFound
	}

	return nil
}
