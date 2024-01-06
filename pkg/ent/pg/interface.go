package pg

import (
	"context"

	"github.com/gnames/gnames/pkg/ent/verif"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnquery/ent/search"
)

type PG interface {
	// DataSourcesMap returns a map of all data-sources known to gnames.
	// Keys are data-source IDs.
	DataSourcesMap() map[int]*vlib.DataSource

	// MatchRecordsMap takes a query input, results of a matching
	// split by type (no match, canonical match, virus match) and
	// returns a map of MatchRecords were keys are input name-strings.
	MatchRecordsMap(
		context.Context,
		verif.MatchSplit,
		vlib.Input) (map[string]*verif.MatchRecord, error)

	// NameByID finds a name-string in the database by its ID.
	// It returns all matches for the name-string accoring to
	// NameStringInput settings. It can limit results to the best match only,
	// it can also filter results by data-sources.
	NameByID(vlib.NameStringInput) (*verif.MatchRecord, error)

	// NameStringByID finds a name-string in the database by its ID.
	// It returns the name-string.
	NameStringByID(string) (string, error)

	// SearchRecordsMap function finds records that correspond to a given
	// advanced search input. It returns a map of MatchRecords were keys
	// are input name-strings.
	SearchRecordsMap(
		ctx context.Context,
		input search.Input,
		spWordIDs []int,
		spWord string) (map[string]*verif.MatchRecord, error)
}
