// Package score allows to sort matched records according to their relevancy.
package score

import (
	"fmt"

	"github.com/gnames/gnames/ent/verifier"
	vlib "github.com/gnames/gnlib/ent/verifier"
)

// Score inteface implements methods for finding best-matching record,
// according to its score, as well as return best matching records from the
// preferred data-sources.
type Score interface {
	fmt.Stringer
	// SortResults takes a pointer for verifier.MatchRecord calculates the score
	// for all MatachResults and sorts them in reverse order according to
	// the calculated score. The highest-scored record is the first, and the
	// lowest-scored record is the last.
	SortResults(*verifier.MatchRecord)
	// BestResult returns the pointer to the best result according to score
	// altorithm.
	BestResult(*verifier.MatchRecord) *vlib.ResultData
	// PreferredResults returns the best-scoring result for each of the
	// given preferred data-sources.
	PreferredResults(sources []int, mr *verifier.MatchRecord) []*vlib.ResultData
}
