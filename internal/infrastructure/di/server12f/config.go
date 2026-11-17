package server12f

import (
	"github.com/spf13/pflag"
	"github.com/tombenke/go-12f-common/v2/config"
	"go.uber.org/multierr"
)

// The configuration parameters of the application.
type Config struct {
}

// Add application-specific config parameters to flagset.
func (cfg *Config) GetConfigFlagSet(fs *pflag.FlagSet) {
}

func (cfg *Config) LoadConfig(fs *pflag.FlagSet) error {
	return multierr.Combine()
}

var _ config.Configurer = (*Config)(nil)
