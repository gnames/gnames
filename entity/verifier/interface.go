// Package verifier provides data about biodiversity
// data-sources and metadata about scientific names aggregated from them.
// The package includes an interface for the data access.
package verifier

import (
	mlib "github.com/gnames/gnlib/domain/entity/matcher"
	vlib "github.com/gnames/gnlib/domain/entity/verifier"
)

// Verifier is an interface that can be implemented by any data provider
// able to prepare raw data for verification.
type Verifier interface {
	// DataSources returns a slice of all data-sources known to gnames. If
	// id are provided, it returns a slice of requested data-sources.
	DataSources(ids ...int) ([]*vlib.DataSource, error)

	// MatchRecords function returns unsorted records corresponding to Input
	// matches.  Matches contain an input name-string, and strings that matched
	// that input.
	MatchRecords(matches []*mlib.Match) (map[string]*MatchRecord, error)
}
