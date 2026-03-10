package configs

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Load loads configuration from the specified file path.
// Environment variables with the STUDIO_ prefix override any value in the file.
//
// Usage:
//   cfg, err := configs.Load("configs/config.local.yaml")  // local dev
//   cfg, err := configs.Load("configs/config.prod.yaml")   // production
func Load(file string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(file)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file %q: %w", file, err)
	}

	// STUDIO_* environment variables override file values.
	// e.g. STUDIO_DATABASE_DSN overrides database.dsn
	v.SetEnvPrefix("STUDIO")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
