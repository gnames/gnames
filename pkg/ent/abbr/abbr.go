// Package abbr finds candidate genera for a group of names with the same
// abbreviated genus, such as `C. elegans` and `C. briggsae`. It parses
// input names, matches them to names from the `abbr_lookups` table,
// groups matches by genus and scores the genera.
//
// The API contract (input and output) lives in gnlib `ent/abbr`. This
// package keeps internal entities, defaults and score weights, so they
// can change without a gnlib release.
package abbr

import (
	"context"
)

// Store provides data from `abbr_lookups` and `abbr_genera` tables.
type Store interface {
	// AbbrLookups returns rows of `abbr_lookups` for the given keys.
	AbbrLookups(ctx context.Context, keys []string) ([]Lookup, error)

	// AbbrGenera returns rows of `abbr_genera` for the given genera.
	// A homonym has several rows, one for each kingdom.
	AbbrGenera(ctx context.Context, genera []string) ([]GenusRecord, error)

	// AbbrPrefixSize returns the sum of sizes (distinct epithet stems) of
	// all genera that start with the given prefix.
	AbbrPrefixSize(ctx context.Context, prefix string) (int, error)

	// AbbrCanonicals returns canonical forms for the given IDs of
	// stemmed canonical forms. Keys of the map are the stem IDs.
	AbbrCanonicals(
		ctx context.Context,
		stemIDs []string,
	) (map[string][]string, error)
}

// Lookup is a row of the `abbr_lookups` table: one epithet of a stemmed
// canonical form.
type Lookup struct {
	// Key is the lowercase first letter of the genus and the epithet stem,
	// for example `c-elegans`.
	Key string

	// Genus is the genus of the name.
	Genus string

	// CanonicalStemID is the ID of the stemmed canonical form.
	CanonicalStemID string

	// CanonicalStem is the stemmed canonical form, for example
	// `Panthera tigr altaic`.
	CanonicalStem string

	// DsNum is the number of data sources with the name.
	DsNum int
}

// GenusRecord is a row of the `abbr_genera` table.
type GenusRecord struct {
	// Genus is the name of the genus.
	Genus string

	// Kingdom of the genus, empty if unknown.
	Kingdom string

	// EpithetsNum is the size of the genus in distinct epithet stems.
	EpithetsNum int

	// NamesNum is the number of records of the genus in the kingdom.
	NamesNum int

	// Classification is the classification path, ending with the genus.
	// It is empty if the genus has no classification.
	Classification string

	// ClassificationRanks are the ranks of the Classification.
	ClassificationRanks string

	// DataSourceID is the source of the classification, 0 if there is
	// none.
	DataSourceID int
}
