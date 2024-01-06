package verif

import (
	"context"

	mlib "github.com/gnames/gnlib/ent/matcher"
	vlib "github.com/gnames/gnlib/ent/verifier"
)

type Verifier interface {

	// DataSources takes a list of data-source IDs and returns a slice of
	// data-sources that correspond to these IDs. If no ids are provided, return
	// all data-sources.
	DataSources(ids ...int) []*vlib.DataSource

	// MatchRecords function returns unsorted records corresponding to Input
	// matches.  Matches contain an input name-string, and strings that matched
	// that input.
	MatchRecords(
		ctx context.Context,
		matches []mlib.Match,
		input vlib.Input,
	) (map[string]*MatchRecord, error)

	// NameByID takes a name-string UUID with options and returns back
	// matched results or an error in case of a failure.
	NameByID(vlib.NameStringInput) (*MatchRecord, error)

	// NameStringByID takes UUID as an argument and returns back a name-string
	// that corresponds to that UUID.
	NameStringByID(string) (string, error)
}
