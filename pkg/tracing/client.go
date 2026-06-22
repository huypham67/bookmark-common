package tracing

import (
	"github.com/newrelic/go-agent/v3/newrelic"
	"github.com/rs/zerolog/log"
)

// NewApplication initializes a New Relic APM application from environment variables.
func NewApplication(envPrefix string) (*newrelic.Application, error) {
	cfg, err := LoadConfig(envPrefix)
	if err != nil {
		return nil, err
	}

	app, err := newrelic.NewApplication(
		newrelic.ConfigAppName(cfg.AppName),
		newrelic.ConfigLicense(cfg.LicenseKey),
		newrelic.ConfigDistributedTracerEnabled(true),
	)
	if err != nil {
		return nil, err
	}

	log.Info().Str("app_name", cfg.AppName).Msg("New Relic APM initialized")
	return app, nil
}
