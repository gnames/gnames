package gnames

import (
	"context"

	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnlib/ent/gnvers"
	"github.com/gnames/gnlib/ent/reconciler"
	"github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnquery/ent/search"
)

// GNames is the main use-case interface of the app. Its purpose to provide
// metadata of registered DataSources and provide functionality for
// verification (resolution/reconciliation) of name-strings to
// known to gnames scientific names, as well as providing data where
// these names occur.
type GNames interface {
	// Verify takes a slice of name-strings together with query parameters and
	// returns back results of verification.
	Verify(ctx context.Context, params verifier.Input) (verifier.Output, error)

	// Reconcile takes the result of verification and converts it into
	// lexical reconciliation groups.
	Reconcile(
		verif verifier.Output,
		qs map[string]reconciler.Query,
		ids []string,
	) reconciler.Output

	// ExtendReconcile takes an Extension query according to
	// Reconciliation Service API and returns back the
	// result according to the API corresponding schema.
	ExtendReconcile(
		reconciler.ExtendQuery,
	) (reconciler.ExtendOutput, error)

	// Search finds scientific names that match the provided partial
	// information. For example, it can handle cases where the genus is
	// abbreviated or only part of the specific epithet is known.
	// It can also utilize year and year range information to narrow
	// down the search.
	Search(ctx context.Context, srch search.Input) search.Output

	// NameByID finds a name-string according to its UUID or exact spelling.
	// The boolean argument allows to return not only identical strings, but
	// all strings that match name-string connected to the ID.
	NameByID(verifier.NameStringInput, bool) (verifier.NameStringOutput, error)

	// Datasources take IDs of data-sourses and return back list of
	// corresponding metadata. If no IDs are given, it returns metadata for all
	// data-sources.
	DataSources(ids ...int) []*verifier.DataSource

	// GetConfig returns configuration of the GNames object.
	GetConfig() config.Config

	// GetVersion returns the version of GNames and a timestamp of its build.
	GetVersion() gnvers.Version
}
