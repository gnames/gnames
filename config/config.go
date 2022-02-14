package config

import (
	"path/filepath"
	"regexp"

	"github.com/gnames/gnsys"
	"github.com/rs/zerolog/log"
)

// Config collects and stores external configuration data.
type Config struct {
	GNport      int
	WorkDir     string
	JobsNum     int
	MaxEditDist int
	PgHost      string
	PgPort      int
	PgUser      string
	PgPass      string
	PgDB        string
	MatcherURL  string
	// NsqdTCPAddress provides an address to the NSQ messenger TCP service. If
	// this value is set and valid, the web logs will be published to the NSQ.
	// The option is ignored if `Port` is not set.
	//
	// If WithWebLogs option is set to `false`, but `NsqdTCPAddress` is set to a
	// valid URL, the logs will be sent to the NSQ messanging service, but they
	// wil not appear as STRERR output.
	// Example: `127.0.0.1:4150`
	NsqdTCPAddress string

	// NsqdContainsFilter logs should match the filter to be sent to NSQ
	// service.
	// Examples:
	// "api" - logs should contain "api"
	// "!api" - logs should not contain "api"
	NsqdContainsFilter string

	// NsqdRegexFilter logs should match the regular expression to be sent to
	// NSQ service.
	// Example: `api\/v(0|1)`
	NsqdRegexFilter *regexp.Regexp

	// WithWebLogs flag enables logs when running web-service. This flag is
	// ignored if `Port` value is not set.
	WithWebLogs bool
}

// NewConfig is a Config constructor that takes external options to
// update default values to external ones.
func NewConfig(opts ...Option) Config {
	workDir, _ := gnsys.ConvertTilda("~/.local/share/gnames")
	cnf := Config{
		GNport:      8888,
		WorkDir:     workDir,
		JobsNum:     8,
		MaxEditDist: 1,
		PgHost:      "localhost",
		PgPort:      5432,
		PgUser:      "postgres",
		PgPass:      "",
		PgDB:        "gnames",
		MatcherURL:  "https://matcher.globalnames.org/api/v1/",
	}
	for _, opt := range opts {
		opt(&cnf)
	}
	return cnf
}

// TrieDir returns path where to dump/restore
// serialized trie.
func (cnf Config) TrieDir() string {
	return filepath.Join(cnf.WorkDir, "levenshein")
}

// FiltersDir returns path where to dump/restore
// serialized bloom filters.
func (cnf Config) FiltersDir() string {
	return filepath.Join(cnf.WorkDir, "bloom")
}

// StemsDir returns path where stems key-value store
// is located
func (cnf Config) StemsDir() string {
	return filepath.Join(cnf.WorkDir, "stems-kv")
}

// Option is a type of all options for Config.
type Option func(cnf *Config)

// OptGNPort sets port for gnames HTTP service.
func OptGNPort(i int) Option {
	return func(cnf *Config) {
		cnf.GNport = i
	}
}

// OptWorkDir sets a directory for key-value stores and temporary files.
func OptWorkDir(s string) Option {
	return func(cnf *Config) {
		cnf.WorkDir, _ = gnsys.ConvertTilda(s)
	}
}

// OptJobsNum sets number of concurrent jobs to run for parallel tasks.
func OptJobsNum(i int) Option {
	return func(cnf *Config) {
		cnf.JobsNum = i
	}
}

// OptMaxEditDist sets maximal possible edit distance for fuzzy matching of
// stemmed canonical forms.
func OptMaxEditDist(i int) Option {
	return func(cnf *Config) {
		if i < 1 || i > 2 {
			log.Warn().Msgf("MaxEditDist can only be 1 or 2, leaving it at %d.",
				cnf.MaxEditDist)
		} else {
			cnf.MaxEditDist = i
		}
	}
}

// OptPgHost sets the host of gnames database
func OptPgHost(s string) Option {
	return func(cnf *Config) {
		cnf.PgHost = s
	}
}

// OptPgUser sets the user of gnnames database
func OptPgUser(s string) Option {
	return func(cnf *Config) {
		cnf.PgUser = s
	}
}

// OptPgPass sets the password to access gnnames database
func OptPgPass(s string) Option {
	return func(cnf *Config) {
		cnf.PgPass = s
	}
}

// OptPgPort sets the port for gnames database
func OptPgPort(i int) Option {
	return func(cnf *Config) {
		cnf.PgPort = i
	}
}

// OptPgDB sets the name of gnames database
func OptPgDB(s string) Option {
	return func(cnf *Config) {
		cnf.PgDB = s
	}
}

// OptMatcherURL sets the name of gnames database
func OptMatcherURL(s string) Option {
	return func(cnf *Config) {
		cnf.MatcherURL = s
	}
}

// OptNsqdTCPAddress provides a URL to NSQ messanging service.
func OptNsqdTCPAddress(s string) Option {
	return func(cfg *Config) {
		cfg.NsqdTCPAddress = s
	}
}

// OptNsqdContainsFilter provides a filter for logs sent to NSQ service.
func OptNsqdContainsFilter(s string) Option {
	return func(cfg *Config) {
		cfg.NsqdContainsFilter = s
	}
}

// OptNsqdRegexFilter provides a regular expression filter for
// logs sent to NSQ service.
func OptNsqdRegexFilter(s string) Option {
	return func(cfg *Config) {
		r := regexp.MustCompile(s)
		cfg.NsqdRegexFilter = r
	}
}

// OptWithWebLogs sets the WithWebLogs field.
func OptWithWebLogs(b bool) Option {
	return func(cfg *Config) {
		cfg.WithWebLogs = b
	}
}
