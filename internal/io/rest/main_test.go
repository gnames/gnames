package rest_test

import (
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gnames/gnames/internal/io/pgio"
	"github.com/gnames/gnames/internal/io/rest"
	"github.com/gnames/gnames/internal/io/srchio"
	"github.com/gnames/gnames/internal/io/verifio"
	"github.com/gnames/gnames/internal/io/vernio"
	gnames "github.com/gnames/gnames/pkg"
	"github.com/gnames/gnames/pkg/config"
)

// testDB is the database for tests. The production `gnames` database is
// never used, whatever the environment says.
const testDB = "gntest"

// restURL is the API URL of the in-process test server.
var restURL string

// TestMain runs tests against an in-process server that uses the test
// database, an embedded matcher and a separate cache, built from the test
// database. PostgreSQL connection settings come from the environment.
func TestMain(m *testing.M) {
	ts, err := testServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot start test server: %v\n", err)
		os.Exit(1)
	}
	restURL = ts.URL + "/api/v1/"
	code := m.Run()
	ts.Close()
	os.Exit(code)
}

func testServer() (*httptest.Server, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	cfg := config.New()
	config.LoadEnv(&cfg)
	for _, opt := range []config.Option{
		config.OptPgDB(testDB),
		config.OptMatcherURL(""),
		config.OptWorkDir(filepath.Join(cacheDir, "gnames-test")),
	} {
		opt(&cfg)
	}

	db, err := pgio.New(cfg)
	if err != nil {
		return nil, err
	}
	vf, err := verifio.New(cfg, db)
	if err != nil {
		return nil, err
	}
	srch, err := srchio.New(cfg, db)
	if err != nil {
		return nil, err
	}
	gn, err := gnames.New(cfg, vf, vernio.New(cfg, db), srch, db)
	if err != nil {
		return nil, err
	}
	return httptest.NewServer(rest.New(gn)), nil
}
