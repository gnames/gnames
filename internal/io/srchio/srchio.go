package srchio

import (
	"context"
	"errors"

	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnames/pkg/ent/pg"
	"github.com/gnames/gnames/pkg/ent/srch"
	"github.com/gnames/gnames/pkg/ent/verif"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnparser/ent/parsed"
	"github.com/gnames/gnquery/ent/search"
)

type srchio struct {
	db pg.PG
	search.Input
	dsm map[int]*vlib.DataSource
}

func New(cnf config.Config, db pg.PG) (srch.Searcher, error) {
	res := srchio{
		db:  db,
		dsm: db.DataSourcesMap(),
	}
	return &res, nil
}

// AdvancedSearch takes a search.Input, perfomes an advanced search
// and returns a map of MatchRecords.
func (s *srchio) AdvancedSearch(
	ctx context.Context,
	input search.Input,
) (map[string]*verif.MatchRecord, error) {
	var err error
	res := make(map[string]*verif.MatchRecord)
	s.Input = input
	spWordIDs, spWord := s.spInput()
	if spWordIDs == nil {
		return res, errors.New("cannot run search without species epithet data")
	}

	res, err = s.db.SearchRecordsMap(ctx, s.Input, spWordIDs, spWord)
	if err != nil {
		return res, err
	}
	return res, nil

}

func (s *srchio) spInput() ([]int, string) {
	if s.SpeciesInfra != "" {
		return []int{int(parsed.InfraspEpithetType)}, s.SpeciesInfra
	} else if s.SpeciesAny != "" {
		return []int{int(parsed.InfraspEpithetType), int(parsed.SpEpithetType)}, s.SpeciesAny
	} else if s.Species != "" {
		return []int{int(parsed.SpEpithetType)}, s.Species
	}
	return nil, ""
}
