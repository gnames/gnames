// Package verifierpg is a PostgreSQL-based implementaion of the Verifier interface.
package verifierpg

import (
	"database/sql"
	"log/slog"

	"github.com/gnames/gnames/internal/io/dbshare"
	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnames/pkg/ent/verifier"
	vlib "github.com/gnames/gnlib/ent/verifier"

	// postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type verifierpg struct {
	db             *sql.DB
	dataSourcesMap map[int]*vlib.DataSource
}

// New creates a new instance of sqlx.DB using configuration data.
func New(cfg config.Config) (verifier.Verifier, error) {
	db, err := sql.Open("postgres", dbshare.DBURL(cfg))
	if err != nil {
		slog.Error("Cannot create PostgreSQL connection", "error", err)
		return verifierpg{}, err
	}
	vf := verifierpg{db: db}
	vf.dataSourcesMap, err = dbshare.DataSourcesMap(db)
	return vf, err
}

func (vf verifierpg) DataSources(ids ...int) ([]*vlib.DataSource, error) {
	return dbshare.DataSources(vf.db, ids...)
}
