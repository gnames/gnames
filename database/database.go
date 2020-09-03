// Package database is an interface to PostgreSQL database that contains Global
// Names data
package database

import (
	"database/sql"
	"fmt"

	"github.com/gnames/gnames/config"
	log "github.com/sirupsen/logrus"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// NewDB creates a new instance of sql.DB using configuration data.
func NewDB(cnf config.Config) *sql.DB {
	db, err := sql.Open("postgres", opts(cnf))
	if err != nil {
		log.Fatalf("Cannot create PostgreSQL connection: %s.", err)
	}
	return db
}

func opts(cnf config.Config) string {
	return fmt.Sprintf("host=%s user=%s port=%d password=%s dbname=%s sslmode=disable",
		cnf.PgHost, cnf.PgUser, cnf.PgPort, cnf.PgPass, cnf.PgDB)
}
