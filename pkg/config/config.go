package config

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gnames/gnsys"
)

// Config collects and stores external configuration data.
type Config struct {
	// Port is the REST service port
	Port int

	// CacheDir is a directory to keep cached data.
	CacheDir string

	// JobsNum is the number of processes to run concurrently.
	JobsNum int

	// MaxEditDist is the the muximum number of Levenschein edits before
	// aborging a fuzzy matching.
	MaxEditDist int

	// MatcherURL is the URL where GNmatcher REST service resides.
	// It is used to get source-agnostic name-matching.
	MatcherURL string

	// WebPageURL provides URL which is used to access GNverifier Web GUI.
	// This URL is used by reconciliation service to create URLs for
	// name-string link and widget.
	WebPageURL string

	// GnamesHostURL provides base URL of the Gnames service. This URL for
	// official service would be the same as WebPageURL. All path information
	// (/api/v?) is stripped off.
	GnamesHostURL string

	// PgDB is the name of GNames database.
	PgDB string

	// PgHost is the domain name or IP address of PostgreSQL service.
	PgHost string

	// PgPass is the password for the PostgreSQL user.
	PgPass string

	// PgPort is the port used by PostgreSQL service.
	PgPort int

	// PgUser is the PostgreSQL user with access to GNames database.
	PgUser string
}

// TrieDir returns path where to dump/restore
// serialized trie.
func (cnf Config) TrieDir() string {
	return filepath.Join(cnf.CacheDir, "levenshein")
}

// FiltersDir returns path where to dump/restore
// serialized bloom filters.
func (cnf Config) FiltersDir() string {
	return filepath.Join(cnf.CacheDir, "bloom")
}

// StemsDir returns path where stems key-value store
// is located
func (cnf Config) StemsDir() string {
	return filepath.Join(cnf.CacheDir, "stems-kv")
}

// Option is a type of all options for Config.
type Option func(cnf *Config)

// OptGNPort sets port for gnames HTTP service.
func OptGNPort(i int) Option {
	return func(cnf *Config) {
		cnf.Port = i
	}
}

// OptWorkDir sets a directory for key-value stores and temporary files.
func OptWorkDir(s string) Option {
	return func(cnf *Config) {
		cnf.CacheDir, _ = gnsys.ConvertTilda(s)
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
			slog.Warn(
				"MaxEditDist entered value is too high, using allowed max",
				slog.Int("entered_val", i),
				slog.Int("max_val", cnf.MaxEditDist),
			)
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

// OptMatcherURL sets the URL to the gnmatcher service.
func OptMatcherURL(s string) Option {
	return func(cnf *Config) {
		cnf.MatcherURL = s
	}
}

// OptWebPageURL sets the URL to the web user interface of gnverifier.
func OptWebPageURL(s string) Option {
	return func(cnf *Config) {
		cnf.WebPageURL = s
	}
}

// OptGnamesHostURL sets the host URL to the gnames service.
func OptGnamesHostURL(s string) Option {
	return func(cnf *Config) {
		cnf.GnamesHostURL = s
	}
}

// New is a Config constructor that takes options to
// update default values.
func New(opts ...Option) Config {
	cacheDir, _ := os.UserCacheDir()
	cnf := Config{
		CacheDir:      filepath.Join(cacheDir, "gnames"),
		JobsNum:       8,
		MatcherURL:    "https://matcher.globalnames.org/api/v1/",
		WebPageURL:    "https://verifier.globalnames.org",
		GnamesHostURL: "https://verifier.globalnames.org",
		MaxEditDist:   1,
		PgDB:          "gnames",
		PgHost:        "0.0.0.0",
		PgPass:        "postgres",
		PgPort:        5432,
		PgUser:        "postgres",
		Port:          8888,
	}

	for _, opt := range opts {
		opt(&cnf)
	}
	return cnf
}
