package verif

import (
	mlib "github.com/gnames/gnlib/ent/matcher"
	vlib "github.com/gnames/gnlib/ent/verifier"
)

// MatchRecord contains information necessary for generating final
// verification output. Most of its fields has the same semantic meaning as
// `entity.Verification` fields.
type MatchRecord struct {
	ID                 string
	Name               string
	Cardinality        int
	CanonicalSimple    string
	CanonicalFull      string
	Authors            []string
	Year               int
	DataSourcesNum     int
	DataSourcesDetails []vlib.DataSourceDetails
	Overload           bool
	// MatchResults contains all matches to Input.
	MatchResults []*vlib.ResultData
	// Sorted indicates if MatchResults are already sorted by their Score field.
	Sorted bool
}

// MatchSplit contains three slices of matches: no match, virus, and canonical.
// They correspond to 3 possible matches types.
type MatchSplit struct {
	// NoMatch contains failed matches.
	NoMatch []*mlib.Match

	// Virus contains matches to virus names.
	Virus []*mlib.Match

	// Canonical contains matches to canonical forms of names.
	Canonical []*mlib.Match
}
