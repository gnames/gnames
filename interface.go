package gnames

import (
	"context"

	"github.com/gnames/gnlib/ent/gnvers"
	vlib "github.com/gnames/gnlib/ent/verifier"
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
	Verify(ctx context.Context, params vlib.VerifyParams) (vlib.Verification, error)
	// Datasources take IDs of data-sourses and return back list of corresponding
	// metadata. If no IDs are given, it returns metadata for all data-sources.
	DataSources(ids ...int) ([]*vlib.DataSource, error)
}
