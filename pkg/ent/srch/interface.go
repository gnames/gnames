package srch

import (
	"context"

	"github.com/gnames/gnames/pkg/ent/verif"
	"github.com/gnames/gnquery/ent/search"
)

// Searcher is an interface that provides methods to do advanced search
// queries.
type Searcher interface {
	// AdvancedSearch finds scientific names that match the provided partial
	// information. For example, it can handle cases where the genus is
	// abbreviated or only part of the specific epithet is known.
	// It can also utilize year and year range information to narrow
	// down the search.
	AdvancedSearch(
		ctx context.Context,
		inp search.Input,
	) (map[string]*verif.MatchRecord, error)
}
