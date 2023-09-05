// Package verifierpg is a PostgreSQL-based implementaion of the Verifier interface.
package verifierpg

import (
	"database/sql"

	"github.com/gnames/gnames/internal/io/dbshare"
	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnames/pkg/ent/verifier"
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
func New(cfg config.Config) verifier.Verifier {
	db, err := sql.Open("postgres", dbshare.DBURL(cfg))
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
