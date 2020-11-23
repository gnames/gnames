package verifier

import (
	vlib "github.com/gnames/gnlib/domain/entity/verifier"
)

// MatchRecord contains information necessary for generating final
// verification output. Most of its fields has the same semantic meaning as
// `entity.Verification` fields.
type MatchRecord struct {
	InputID         string
	Input           string
	Cardinality     int
	CanonicalSimple string
	CanonicalFull   string
	Authors         []string
	Year            int
	MatchType       vlib.MatchTypeValue
	Curation        vlib.CurationLevel
	DataSourcesNum  int
	// MatchResults contains all matches to Input.
	MatchResults []*vlib.ResultData
	// Sorted indicates if MatchResults are already sorted by their Score field.
	Sorted bool
}
