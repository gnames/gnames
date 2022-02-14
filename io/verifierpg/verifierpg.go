// Package verifierpg is a PostgreSQL-based implementaion of the Verifier interface.
package verifierpg

import (
	"database/sql"
	"fmt"

	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/ent/verifier"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/rs/zerolog/log"

	// postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type verifierpg struct {
	DB             *sql.DB
	DataSourcesMap map[int]*vlib.DataSource
}

// NewVerifier creates a new instance of sql.DB using configuration data.
func NewVerifier(cnf config.Config) verifier.Verifier {
	db, err := sql.Open("postgres", dbURL(cnf))
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create PostgreSQL connection")
	}
	vf := verifierpg{DB: db}
	vf.dataSourcesMap()
	return vf
}

func dbURL(cnf config.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cnf.PgUser, cnf.PgPass, cnf.PgHost, cnf.PgPort, cnf.PgDB)
}

func (vf *verifierpg) dataSourcesMap() {
	dsm := make(map[int]*vlib.DataSource)
	dss, err := vf.DataSources()
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot init DataSources data")
	}
	for _, ds := range dss {
		dsm[ds.ID] = ds
	}
	vf.DataSourcesMap = dsm
}
