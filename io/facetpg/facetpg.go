package facetpg

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gnames/gnames/config"
	"github.com/gnames/gnames/ent/facet"
	"github.com/gnames/gnames/ent/verifier"
	"github.com/gnames/gnames/io/internal/dbshare"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnparser/ent/parsed"
	"github.com/gnames/gnquery/ent/search"
	"github.com/rs/zerolog/log"
)

type facetpg struct {
	db *sql.DB
	search.Input
	spWord    string
	spWordIDs []int
	dsm       map[int]*vlib.DataSource
}

func New(cnf config.Config) facet.Facet {
	dbURL := dbshare.DBURL(cnf)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal().Err(err).Msg("Cannot create PostgreSQL connection")
	}
	return &facetpg{db: db, dsm: dbshare.DataSourcesMap(db)}
}

func (f *facetpg) Search(
	ctx context.Context,
	inp search.Input,
) (map[string]*verifier.MatchRecord, error) {
	var err error
	res := make(map[string]*verifier.MatchRecord)
	f.Input = inp
	f.spWordIDs, f.spWord = f.spInput()
	if f.spWordIDs == nil {
		return res, errors.New("cannot run search without species epithet data")
	}
	q, args := f.setQuery()

	res, err = f.runQuery(ctx, q, args)
	if err != nil {
		return res, err
	}
	return res, nil
}

func (f *facetpg) spInput() ([]int, string) {
	if f.SpeciesInfra != "" {
		return []int{int(parsed.InfraspEpithetType)}, f.SpeciesInfra
	} else if f.SpeciesAny != "" {
		return []int{int(parsed.InfraspEpithetType), int(parsed.SpEpithetType)}, f.SpeciesAny
	} else if f.Species != "" {
		return []int{int(parsed.SpEpithetType)}, f.Species
	}
	return nil, ""
}
