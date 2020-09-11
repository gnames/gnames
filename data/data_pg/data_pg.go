// Package data_pg is a PostgreSQL-based implementaion of the data.DataGrabber
// interface.
package data_pg

import (
	"database/sql"
	"fmt"

	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/data"
	"github.com/gnames/gnames/domain/entity"
	log "github.com/sirupsen/logrus"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type DataGrabberPG struct {
	DB             *sql.DB
	DataSourcesMap map[int]*entity.DataSource
}

// NewDB creates a new instance of sql.DB using configuration data.
func NewDataGrabberPG(cnf config.Config) DataGrabberPG {
	db, err := sql.Open("postgres", opts(cnf))
	if err != nil {
		log.Fatalf("Cannot create PostgreSQL connection: %s.", err)
	}
	dgp := DataGrabberPG{DB: db}
	dgp.dataSourcesMap()
	return dgp
}

func opts(cnf config.Config) string {
	return fmt.Sprintf("host=%s user=%s port=%d password=%s dbname=%s sslmode=disable",
		cnf.PgHost, cnf.PgUser, cnf.PgPort, cnf.PgPass, cnf.PgDB)
}

func (dgp *DataGrabberPG) dataSourcesMap() {
	dsm := make(map[int]*entity.DataSource)
	dss, err := dgp.DataSources(data.NullInt{})
	if err != nil {
		log.Fatalf("Cannot init DataSources data: %s", err)
	}
	for _, ds := range dss {
		dsm[ds.ID] = ds
	}
	dgp.DataSourcesMap = dsm
}
