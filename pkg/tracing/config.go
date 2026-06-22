package tracing

import "github.com/kelseyhightower/envconfig"

// Config holds APM configuration loaded from environment variables.
type Config struct {
	LicenseKey string `envconfig:"NEW_RELIC_LICENSE_KEY" required:"true"`
	AppName    string `envconfig:"NEW_RELIC_APP_NAME" default:"app"`
}

// LoadConfig loads APM configuration from environment variables with the given prefix.
func LoadConfig(envPrefix string) (*Config, error) {
	cfg := &Config{}
	if err := envconfig.Process(envPrefix, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
