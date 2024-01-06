package verifio

import (
	"cmp"
	"context"
	"log/slog"
	"slices"

	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnames/pkg/ent/pg"
	"github.com/gnames/gnames/pkg/ent/verif"
	mlib "github.com/gnames/gnlib/ent/matcher"
	vlib "github.com/gnames/gnlib/ent/verifier"
)

type verifio struct {
	db  pg.PG
	dsm map[int]*vlib.DataSource
}

func New(cfg config.Config, db pg.PG) (verif.Verifier, error) {
	res := verifio{
		db:  db,
		dsm: db.DataSourcesMap(),
	}
	return &res, nil
}

// DataSources returns an array of data sources based on the provided IDs.
// If no IDs are provided, it returns all data sources.
// If an ID is not found in the data source map, a warning is logged.
// The returned data sources are sorted by their ID in ascending order.
func (v *verifio) DataSources(ids ...int) []*vlib.DataSource {
	res := make([]*vlib.DataSource, 0, len(ids))
	if len(ids) > 0 {
		for _, i := range ids {
			if ds, ok := v.dsm[i]; ok {
				res = append(res, ds)
			} else {
				slog.Warn("Data source not found", "id", i)
			}
		}
		if len(res) > 0 {
			return res
		}
	}

	for _, ds := range v.dsm {
		res = append(res, ds)
	}
	slices.SortFunc(res, func(a, b *vlib.DataSource) int {
		return cmp.Compare(a.ID, b.ID)
	})
	return res
}

// MatchRecords function returns unsorted records corresponding to Input
// matches.  Matches contain an input name-string, and strings that matched
// that input.
func (vrf *verifio) MatchRecords(
	ctx context.Context,
	matches []mlib.Match,
	input vlib.Input,
) (map[string]*verif.MatchRecord, error) {
	var res map[string]*verif.MatchRecord

	// separate NoMatch, Virus, and matches
	splitMatches := partitionMatches(matches)

	res, err := vrf.db.MatchRecordsMap(ctx, splitMatches, input)
	if err != nil {
		slog.Error("Cannot get matches data", "error", err)
		return res, err
	}
	return res, nil
}

// NameByID takes a name-string UUID with options and returns back
// matched results or an error in case of a failure.
func (v *verifio) NameByID(inp vlib.NameStringInput) (*verif.MatchRecord, error) {
	return v.db.NameByID(inp)
}

// NameStringByID takes UUID as an argument and returns back a name-string
// that corresponds to that UUID.
func (v *verifio) NameStringByID(id string) (string, error) {
	return v.db.NameStringByID(id)
}
