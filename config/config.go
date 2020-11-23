package config

import (
	"fmt"
	"path/filepath"

	"github.com/gnames/gnlib/sys"
	log "github.com/sirupsen/logrus"
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
}

// NewConfig is a Config constructor that takes external options to
// update default values to external ones.
func NewConfig(opts ...Option) Config {
	cnf := Config{
		GNport:      8888,
		WorkDir:     sys.ConvertTilda("~/.local/share/gnames"),
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
		cnf.WorkDir = sys.ConvertTilda(s)
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
			log.Warn(fmt.Sprintf("MaxEditDist can only be 1 or 2, leaving it at %d.",
				cnf.MaxEditDist))
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
