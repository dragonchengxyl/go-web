package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/studio/platform/internal/domain/user"
	"github.com/studio/platform/internal/infra/postgres"
	"github.com/studio/platform/internal/pkg/crypto"
)

func newAdminCmd(opts *Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "admin",
		Short: "Bootstrap or promote administrative users",
	}

	var (
		email    string
		username string
		password string
		role     string
	)

	ensureCmd := &cobra.Command{
		Use:   "ensure",
		Short: "Promote an existing user to admin, or create one if it does not exist",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdminEnsure(cmd.Context(), opts, adminEnsureInput{
				Email:    email,
				Username: username,
				Password: password,
				Role:     role,
			})
		},
	}

	ensureCmd.Flags().StringVar(&email, "email", "", "User email")
	ensureCmd.Flags().StringVar(&username, "username", "", "Username for creating a new admin if the email does not exist")
	ensureCmd.Flags().StringVar(&password, "password", "", "Password for creating a new admin if the email does not exist")
	ensureCmd.Flags().StringVar(&role, "role", "admin", "Role to grant: admin or super_admin")
	_ = ensureCmd.MarkFlagRequired("email")

	cmd.AddCommand(ensureCmd)
	return cmd
}

type adminEnsureInput struct {
	Email    string
	Username string
	Password string
	Role     string
}

func runAdminEnsure(ctx context.Context, opts *Options, input adminEnsureInput) error {
	cfg, err := opts.loadConfig(true)
	if err != nil {
		return err
	}

	role, err := parseAdminRole(input.Role)
	if err != nil {
		return err
	}

	email := strings.TrimSpace(strings.ToLower(input.Email))
	if err := crypto.ValidateEmail(email); err != nil {
		return err
	}

	pool, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	repo := postgres.NewUserRepository(pool)

	existing, err := repo.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		now := time.Now()
		previousRole := existing.Role
		existing.Role = role
		existing.Status = user.StatusActive
		existing.ForcePasswordReset = user.NextForcePasswordReset(existing.ForcePasswordReset, previousRole, role)
		if existing.EmailVerifiedAt == nil {
			existing.EmailVerifiedAt = &now
		}
		existing.UpdatedAt = now
		if err := repo.Update(ctx, existing); err != nil {
			return fmt.Errorf("promote existing user: %w", err)
		}
		fmt.Fprintf(opts.Out, "updated existing user %s -> role=%s\n", existing.Email, existing.Role)
		return nil
	}
	if err != nil && err != user.ErrNotFound {
		return fmt.Errorf("lookup user by email: %w", err)
	}

	username := strings.TrimSpace(input.Username)
	if username == "" {
		username = suggestedUsername(email)
	}
	if err := crypto.ValidateUsername(username); err != nil {
		return fmt.Errorf("username %q invalid: %w", username, err)
	}

	password := strings.TrimSpace(input.Password)
	if password == "" {
		return fmt.Errorf("user %s does not exist; provide --password to create an admin account", email)
	}
	if err := crypto.ValidatePassword(password, crypto.DefaultPasswordStrength()); err != nil {
		return err
	}
	if crypto.IsCommonPassword(password) {
		return fmt.Errorf("password is too common")
	}

	existsByUsername, err := repo.ExistsByUsername(ctx, username)
	if err != nil {
		return fmt.Errorf("check username existence: %w", err)
	}
	if existsByUsername {
		return fmt.Errorf("username %q already exists; pass --username with another value", username)
	}

	passwordHash, err := crypto.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	now := time.Now()
	newUser := &user.User{
		ID:                 uuid.New(),
		Username:           username,
		Email:              email,
		PasswordHash:       passwordHash,
		Role:               role,
		Status:             user.StatusActive,
		ForcePasswordReset: user.NextForcePasswordReset(false, "", role),
		EmailVerifiedAt:    &now,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := repo.Create(ctx, newUser); err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}

	fmt.Fprintf(opts.Out, "created admin user %s (username=%s role=%s)\n", newUser.Email, newUser.Username, newUser.Role)
	return nil
}

func parseAdminRole(raw string) (user.Role, error) {
	switch strings.TrimSpace(raw) {
	case "", string(user.RoleAdmin):
		return user.RoleAdmin, nil
	case string(user.RoleSuperAdmin):
		return user.RoleSuperAdmin, nil
	default:
		return "", fmt.Errorf("unsupported role %q; use admin or super_admin", raw)
	}
}

func suggestedUsername(email string) string {
	local := email
	if idx := strings.Index(local, "@"); idx >= 0 {
		local = local[:idx]
	}
	local = strings.ReplaceAll(local, ".", "_")
	local = strings.ReplaceAll(local, "+", "_")
	local = strings.ReplaceAll(local, "-", "_")
	if len(local) > 20 {
		local = local[:20]
	}
	if len(local) < 3 {
		local += "_admin"
	}
	return local
}
