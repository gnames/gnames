package dbshare

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/gnames/gnames/config"
	"github.com/gnames/gnparser/ent/parsed"
)

func DBURL(cnf config.Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cnf.PgUser, cnf.PgPass, cnf.PgHost, cnf.PgPort, cnf.PgDB)
}

// MatchRecords connects result data to input name-string. Input name-string
// is a key.
type VerifSQL struct {
	CanonicalID         sql.NullString
	Canonical           sql.NullString
	CanonicalFull       sql.NullString
	Name                sql.NullString
	Cardinality         int
	RecordID            sql.NullString
	NameStringID        sql.NullString
	DataSourceID        int
	LocalID             sql.NullString
	OutlinkID           sql.NullString
	AcceptedRecordID    sql.NullString
	AcceptedNameID      sql.NullString
	AcceptedName        sql.NullString
	Classification      sql.NullString
	ClassificationRanks sql.NullString
	ClassificationIds   sql.NullString
	ParseQuality        int
}

var QueryFields = `
  v.canonical_id, v.name, v.data_source_id, v.record_id,
  v.name_string_id, v.local_id, v.outlink_id, v.accepted_record_id,
  v.accepted_name_id, v.accepted_name, v.classification,
  v.classification_ranks, v.classification_ids, v.parse_quality
`

// ProcessAuthorship converts year to int and provides authors as a slice.
func ProcessAuthorship(au *parsed.Authorship) ([]string, int) {
	authors := make([]string, 0, 2)
	var year int
	if au == nil {
		return authors, year
	}

	authors = au.Authors

	year, err := strconv.Atoi(au.Year)
	if err == nil && !au.Original.Year.IsApproximate {
		return authors, year
	}

	if au.Combination != nil && au.Combination.Year != nil {
		year, _ = strconv.Atoi(au.Combination.Year.Value)
		if au.Combination.Year.IsApproximate {
			year = 0
		}
	}
	return authors, year
}
