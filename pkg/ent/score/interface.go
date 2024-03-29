// Package score allows to sort matched records according to their relevancy.
package score

import (
	"fmt"

	"github.com/gnames/gnames/pkg/ent/verif"
	vlib "github.com/gnames/gnlib/ent/verifier"
)

// Score inteface implements methods for finding best-matching record,
// according to its score, as well as return best matching records from the
// selected data-sources.
type Score interface {
	fmt.Stringer

	// SortResults takes a pointer for verifier.MatchRecord calculates the score
	// for all MatachResults and sorts them in reverse order according to
	// the calculated score. The highest-scored record is the first, and the
	// lowest-scored record is the last.
	SortResults(*verif.MatchRecord)

	// Results returns the best-scoring result for each of the
	// given selected data-sources.
	Results(mr *verif.MatchRecord) []*vlib.ResultData
}
