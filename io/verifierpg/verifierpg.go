// Package verifierpg is a PostgreSQL-based implementaion of the Verifier interface.
package verifierpg

import (
	"database/sql"

	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/ent/verifier"
	"github.com/gnames/gnames/io/internal/dbshare"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/rs/zerolog/log"

	// postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type verifierpg struct {
	db             *sql.DB
	dataSourcesMap map[int]*vlib.DataSource
}

// New creates a new instance of sqlx.DB using configuration data.
func New(cnf config.Config) verifier.Verifier {
	db, err := sql.Open("postgres", dbshare.DBURL(cnf))
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create PostgreSQL connection")
	}
	vf := verifierpg{db: db}
	vf.dataSourcesMap = dbshare.DataSourcesMap(db)
	return vf
}

func (vf verifierpg) DataSources(ids ...int) ([]*vlib.DataSource, error) {
	return dbshare.DataSources(vf.db, ids...)
}
