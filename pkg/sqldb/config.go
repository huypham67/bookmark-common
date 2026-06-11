package sqldb

import "github.com/kelseyhightower/envconfig"

// Config holds the database connection configuration.
type Config struct {
	Host     string `envconfig:"DB_HOST" default:"localhost"`
	Port     string `envconfig:"DB_PORT" default:"5432"`
	User     string `envconfig:"DB_USER" default:"admin"`
	Password string `envconfig:"DB_PASSWORD" default:"admin"`
	Database string `envconfig:"DB_NAME" default:"bookmark_db"`
	SSLMode  string `envconfig:"DB_SSLMODE" default:"disable"`
	TimeZone string `envconfig:"DB_TIMEZONE" default:"UTC"`
}

// LoadConfig loads database configuration from environment variables with the given prefix.
func LoadConfig(prefix string) (*Config, error) {
	cfg := &Config{}
	err := envconfig.Process(prefix, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
