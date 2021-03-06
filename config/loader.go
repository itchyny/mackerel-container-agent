package config

import (
	"context"
	"reflect"
	"time"

	"github.com/mackerelio/golib/logging"
)

var logger = logging.GetLogger("config")

// Loader represents a config loader
type Loader struct {
	location        string
	pollingDuration time.Duration
	lastConfig      *Config
}

// NewLoader creates a new Loader
func NewLoader(location string, pollingDuration time.Duration) *Loader {
	return &Loader{location: location, pollingDuration: pollingDuration}
}

// Load agent configuration
func (l *Loader) Load(ctx context.Context) (*Config, error) {
	config, err := load(ctx, l.location)
	if err != nil {
		return nil, err
	}
	l.lastConfig = config
	return config, nil
}

// Start the loader loop
func (l *Loader) Start(ctx context.Context) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		if l.pollingDuration > 0 {
			t := time.NewTicker(l.pollingDuration)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					config, err := load(ctx, l.location)
					if err != nil {
						logger.Warningf("failed to load config: %s", err)
					} else if !reflect.DeepEqual(l.lastConfig, config) {
						logger.Infof("detected config changes")
						return
					}
				}
			}
		} else {
			<-ctx.Done()
		}
	}()
	return ch
}
