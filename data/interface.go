// Package data provides metadata about biodiversity
// data-sources and metadata about scientific names aggregated from them.
// The package includes an interface for interaction with the data and
// an implementation of the interface for PostgreSQL database.
package data

import (
	gn "github.com/gnames/gnames/domain/entity"
	gnm "github.com/gnames/gnmatcher/domain/entity"
)

// MatchRecord contains information necessary for generating final
// verification output. Most of its fields has the same semantic meaning as
// `entity.Verification` fields.
type MatchRecord struct {
	InputID         string
	Input           string
	Cardinality     int
	CanonicalSimple string
	CanonicalFull   string
	Authors         []string
	MatchType       gn.MatchType
	CurationLevel   gn.CurationLevel
	DataSourcesNum  int
	// MatchResults contains all matches to Input.
	MatchResults []*gn.ResultData
	// Sorted indicates if MatchResults are already sorted by their Score field.
	Sorted bool
}

// An implemenation of int that can be 'nil'.
type NullInt struct {
	Int   int
	Valid bool
}

// DataGrabber is an interface that can be implemented by any data provider
// able to prepare raw data for verification.
type DataGrabber interface {
	// DataSrouces returns a slice of all data-sources known to gnames. If
	// id argument is not nil, it returns a slice with atmost one data-source
	// founc by its id.
	DataSources(id NullInt) ([]*gn.DataSource, error)

	// MatchRecords function returns unsorted records corresponding to Input
	// matches.  Matches contain an input name-string, and strings that matched
	// that input.
	MatchRecords(matches []*gnm.Match) (map[string]*MatchRecord, error)
}
