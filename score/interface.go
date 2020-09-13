package score

import (
	"fmt"

	"github.com/gnames/gnames/data"
	"github.com/gnames/gnames/domain/entity"
)

// Scorer is an interface of a scoring system.
type Scorer interface {
	// Calculate compares input name-string with the matching ResultData and
	// decides on a numerical representation of the result. This result is then
	// used to sort records to define the best matches.
	Calculate(*data.MatchRecord, *entity.ResultData) uint32

	// Sorts ResultData recorda according to score.
	Sort(*data.MatchRecord)

	// BestResult determines the best match out of all results.
	BestResult(*data.MatchRecord) *entity.ResultData

	// Determines the best match for all preferred data sources and aggretates
	// them into a collection sorted by data-source IDs.
	// `sources` is a slice of DataSourceIDs, `mr` is a MatchRecord that
	// contains all ResultData found by matching.
	PreferredResults(sources []int, mr *data.MatchRecord) []*entity.ResultData

	fmt.Stringer
}
