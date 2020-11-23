package config_test

import (
	"testing"

	"github.com/gnames/gnames/config"
	"github.com/gnames/gnlib/sys"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	cnf := config.NewConfig()
	deflt := config.Config{
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
	assert.Equal(t, cnf, deflt)
}

func TestNewOpts(t *testing.T) {
	opts := opts()
	cnf := config.NewConfig(opts...)
	updt := config.Config{
		GNport:      8888,
		WorkDir:     sys.ConvertTilda("~/.local/share/gnames"),
		JobsNum:     16,
		MaxEditDist: 2,
		PgHost:      "mypg",
		PgPort:      1234,
		PgUser:      "gnm",
		PgPass:      "secret",
		PgDB:        "gnm",
		MatcherURL:  "https://matcher.globalnames.org/api/v1/",
	}
	assert.Equal(t, cnf, updt)
}

func TestMaxED(t *testing.T) {
	log.SetLevel(log.FatalLevel)
	cnf := config.NewConfig(config.OptMaxEditDist(5))
	assert.Equal(t, cnf.MaxEditDist, 1)
	cnf = config.NewConfig(config.OptMaxEditDist(0))
	assert.Equal(t, cnf.MaxEditDist, 1)
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
	}
}
