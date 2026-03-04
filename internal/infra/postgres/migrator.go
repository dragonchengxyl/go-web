package postgres

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

// Migrator handles database migrations
type Migrator struct {
	logger *zap.Logger
}

// NewMigrator creates a new Migrator instance
func NewMigrator(logger *zap.Logger) *Migrator {
	return &Migrator{
		logger: logger,
	}
}

// RunMigrations executes all pending database migrations
func (m *Migrator) RunMigrations(databaseURL string, migrationsPath string) error {
	m.logger.Info("Starting database migrations", zap.String("path", migrationsPath))

	// Create migrator instance
	migrator, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	// Get current version
	version, dirty, err := migrator.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("failed to get current version: %w", err)
	}

	if dirty {
		m.logger.Warn("Database is in dirty state", zap.Uint("version", version))
		return fmt.Errorf("database is in dirty state at version %d, manual intervention required", version)
	}

	if err == migrate.ErrNilVersion {
		m.logger.Info("No migrations applied yet, starting from scratch")
	} else {
		m.logger.Info("Current migration version", zap.Uint("version", version))
	}

	// Run migrations
	if err := migrator.Up(); err != nil {
		if err == migrate.ErrNoChange {
			m.logger.Info("No new migrations to apply")
			return nil
		}
		return fmt.Errorf("migration failed: %w", err)
	}

	// Get new version
	newVersion, _, err := migrator.Version()
	if err != nil {
		return fmt.Errorf("failed to get new version: %w", err)
	}

	m.logger.Info("Migrations completed successfully", zap.Uint("new_version", newVersion))
	return nil
}

// Rollback rolls back the last migration
func (m *Migrator) Rollback(databaseURL string, migrationsPath string, steps int) error {
	m.logger.Info("Rolling back migrations", zap.Int("steps", steps))

	migrator, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	if err := migrator.Steps(-steps); err != nil {
		return fmt.Errorf("rollback failed: %w", err)
	}

	m.logger.Info("Rollback completed successfully")
	return nil
}

// GetVersion returns the current migration version
func (m *Migrator) GetVersion(databaseURL string, migrationsPath string) (uint, bool, error) {
	migrator, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrator: %w", err)
	}
	defer migrator.Close()

	version, dirty, err := migrator.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return 0, false, err
	}

	if err == migrate.ErrNilVersion {
		return 0, false, nil
	}

	return version, dirty, nil
}
