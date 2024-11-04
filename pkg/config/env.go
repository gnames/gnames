package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
)

// LoadEnv allows to change configuration during runtime without Cobra CLI.
// Useful for modifying configuration for tests.
func LoadEnv(c *Config) {
	slog.Info("Updating config using environment variables")
	opts := strOpts()
	opts = append(opts, intOpts()...)
	for _, opt := range opts {
		opt(c)
	}
}

func strOpts() []Option {
	var res []Option

	envToOpt := map[string]func(string) Option{
		"GN_MATCHER_URL":     OptMatcherURL,
		"GN_WEB_PAGE_URL":    OptWebPageURL,
		"GN_GNAMES_HOST_URL": OptGnamesHostURL,
		"GN_PG_HOST":         OptPgHost,
		"GN_PG_USER":         OptPgUser,
		"GN_PG_PASS":         OptPgPass,
		"GN_PG_DB":           OptPgDB,
	}

	for envVar, optFunc := range envToOpt {
		envVal := strings.TrimSpace(os.Getenv(envVar))
		if envVal != "" {
			res = append(res, optFunc(envVal))
		}
	}

	return res
}

func intOpts() []Option {
	var res []Option
	envToOpt := map[string]func(int) Option{
		"GN_PORT":          OptGNPort,
		"GN_PG_PORT":       OptPgPort,
		"GN_JOBS_NUM":      OptJobsNum,
		"GN_MAX_EDIT_DIST": OptMaxEditDist,
	}
	for envVar, optFunc := range envToOpt {
		val := strings.TrimSpace(os.Getenv(envVar))
		if val == "" {
			continue
		}
		i, err := strconv.Atoi(val)
		if err != nil {
			slog.Warn("Cannot convert to int", "env", envVar, "value", val)
			continue
		}
		res = append(res, optFunc(i))
	}
	return res
}
