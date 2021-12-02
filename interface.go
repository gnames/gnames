package gnames

import (
	"context"

	"github.com/gnames/gnlib/ent/gnvers"
	"github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnquery/ent/search"
)

// GNames is the main use-case interface of the app. Its purpose to provide
// metadata of registered DataSources and provide functionality for
// verification (resolution/reconciliation) of name-strings to
// known to gnames scientific names, as well as providing data where
// these names occur.
type GNames interface {
	// GetVersion returns the version of GNames and a timestamp of its build.
	GetVersion() gnvers.Version

	// Verify takes a slice of name-strings together with query parameters and
	// returns back results of verification.
	Verify(ctx context.Context, params verifier.Input) (verifier.Output, error)

	// Search takes
	Search(ctx context.Context, srch search.Input) search.Output

	// Datasources take IDs of data-sourses and return back list of corresponding
	// metadata. If no IDs are given, it returns metadata for all data-sources.
	DataSources(ids ...int) ([]*verifier.DataSource, error)
}
