package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnsys"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cnf := config.New()
	cacheDir, _ := os.UserCacheDir()
	deflt := config.Config{
		Port:          8888,
		CacheDir:      filepath.Join(cacheDir, "gnames"),
		JobsNum:       8,
		MaxEditDist:   1,
		PgHost:        "0.0.0.0",
		PgPort:        5432,
		PgUser:        "postgres",
		PgPass:        "postgres",
		PgDB:          "gnames",
		MatcherURL:    "https://matcher.globalnames.org/api/v1/",
		WebPageURL:    "https://verifier.globalnames.org",
		GnamesHostURL: "https://verifier.globalnames.org",
	}
	assert.Equal(t, deflt, cnf)
}

func TestNewOpts(t *testing.T) {
	opts := opts()
	cnf := config.New(opts...)
	workDir, _ := gnsys.ConvertTilda("~/.local/share/gnames")
	updt := config.Config{
		Port:          8888,
		CacheDir:      workDir,
		JobsNum:       16,
		MaxEditDist:   2,
		PgHost:        "mypg",
		PgPort:        1234,
		PgUser:        "gnm",
		PgPass:        "secret",
		PgDB:          "gnm",
		MatcherURL:    "https://matcher.globalnames.org/api/v1/",
		WebPageURL:    "https://example.org",
		GnamesHostURL: "https://example.com",
	}
	assert.Equal(t, updt, cnf)
}

func TestMaxED(t *testing.T) {
	cnf := config.New(config.OptMaxEditDist(5))
	assert.Equal(t, 1, cnf.MaxEditDist)
	cnf = config.New(config.OptMaxEditDist(0))
	assert.Equal(t, 1, cnf.MaxEditDist)
}

func opts() []config.Option {
	return []config.Option{
		config.OptWorkDir("~/.local/share/gnames"),
		config.OptJobsNum(16),
		config.OptMaxEditDist(2),
		config.OptPgHost("mypg"),
		config.OptPgUser("gnm"),
		config.OptPgPass("secret"),
		config.OptPgPort(1234),
		config.OptPgDB("gnm"),
		config.OptWebPageURL("https://example.org"),
		config.OptGnamesHostURL("https://example.com"),
	}
}
