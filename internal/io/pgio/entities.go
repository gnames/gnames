package pgio

import (
	"database/sql"
	"strconv"

	"github.com/gnames/gnparser/ent/parsed"
	"github.com/jackc/pgx/v5"
)

// verifSQL is an intermediate data to create verif.MatchRecords
type verifSQL struct {
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

func rowsToVerifSQL(rows pgx.Rows) ([]*verifSQL, error) {
	var err error
	var res []*verifSQL
	for rows.Next() {
		var v verifSQL
		err = rows.Scan(
			&v.CanonicalID, &v.Name, &v.DataSourceID, &v.RecordID,
			&v.NameStringID, &v.LocalID, &v.OutlinkID, &v.AcceptedRecordID,
			&v.AcceptedNameID, &v.AcceptedName, &v.Classification,
			&v.ClassificationRanks, &v.ClassificationIds, &v.ParseQuality,
		)
		if err != nil {
			return nil, err
		}
		res = append(res, &v)
	}

	return res, nil
}

// processAuthorship converts year to int and provides authors as a slice.
func processAuthorship(au *parsed.Authorship) ([]string, int) {
	authors := make([]string, 0, 2)
	var year int
	if au == nil {
		return authors, year
	}

	authors = au.Authors

	year, _ = strconv.Atoi(au.Year)
	if year > 0 && au.Original != nil &&
		au.Original.Year != nil && !au.Original.Year.IsApproximate {
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
