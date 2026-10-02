package tegrastatsreceiver

import (
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/scraper/scraperhelper"
)

type Mode string

const (
	// ModePush emits metrics on every line `tegrastats` prints,
	// so the rate is set by its `--interval` option.
	ModePush Mode = "push"
	// ModeScrape emits the latest line every `collection_interval`, like `hostmetrics` does.
	ModeScrape Mode = "scrape"
)

type Config struct {
	// Used only in scrape mode.
	scraperhelper.ControllerConfig `mapstructure:",squash"`

	Mode Mode `mapstructure:"mode"`

	// Command to run and its arguments.
	// `["$fake"]` generates fake stats instead of running a command.
	Command []string `mapstructure:"command"`

	// In scrape mode, a line older than this is not emitted.
	StaleTimeout time.Duration `mapstructure:"stale_timeout"`
}

func (c *Config) Validate() error {
	errs := []error{}
	switch c.Mode {
	case ModePush, ModeScrape:
	default:
		errs = append(errs, fmt.Errorf(`"mode" must be %q or %q: %q`, ModePush, ModeScrape, c.Mode))
	}
	if len(c.Command) == 0 || c.Command[0] == "" {
		errs = append(errs, errors.New(`"command" must not be empty`))
	}
	if c.Mode == ModeScrape && c.StaleTimeout <= 0 {
		errs = append(errs, errors.New(`"stale_timeout" must be positive`))
	}
	return errors.Join(errs...)
}
