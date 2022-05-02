// Package verifier provides data about biodiversity
// data-sources and metadata about scientific names aggregated from them.
// The package includes an interface for the data access.
package verifier

import (
	"context"

	mlib "github.com/gnames/gnlib/ent/matcher"
	vlib "github.com/gnames/gnlib/ent/verifier"
)

// Verifier is an interface that can be implemented by any data provider
// able to prepare raw data for verification.
type Verifier interface {
	// DataSources returns a slice of all data-sources known to gnames. If
	// idd are provided, it returns a slice of requested data-sources.
	DataSources(ids ...int) ([]*vlib.DataSource, error)

	// MatchRecords function returns unsorted records corresponding to Input
	// matches.  Matches contain an input name-string, and strings that matched
	// that input.
	MatchRecords(
		ctx context.Context,
		matches []mlib.Output,
		input vlib.Input,
	) (map[string]*MatchRecord, error)
}
