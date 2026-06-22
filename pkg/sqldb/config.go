package sqldb

import "github.com/kelseyhightower/envconfig"

// Config holds the database connection configuration.
type Config struct {
	Host     string `envconfig:"DB_HOST"     required:"true"`
	Port     string `envconfig:"DB_PORT"     default:"5432"`
	User     string `envconfig:"DB_USER"     required:"true"`
	Password string `envconfig:"DB_PASSWORD" required:"true"`
	Database string `envconfig:"DB_NAME"     required:"true"`
	SSLMode  string `envconfig:"DB_SSLMODE"  default:"disable"`
	TimeZone string `envconfig:"DB_TIMEZONE" default:"UTC"`
}

// LoadConfig loads database configuration from environment variables with the given prefix.
func LoadConfig(prefix string) (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process(prefix, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
