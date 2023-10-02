package verifier

import (
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
