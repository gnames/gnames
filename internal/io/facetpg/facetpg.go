package facetpg

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/gnames/gnames/internal/io/dbshare"
	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnames/pkg/ent/facet"
	"github.com/gnames/gnames/pkg/ent/verifier"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnparser/ent/parsed"
	"github.com/gnames/gnquery/ent/search"
)

type facetpg struct {
	db *sql.DB
	search.Input
	spWord    string
	spWordIDs []int
	dsm       map[int]*vlib.DataSource
}

func New(cnf config.Config) (facet.Facet, error) {
	dbURL := dbshare.DBURL(cnf)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		slog.Error("Cannot create PostgreSQL connection", "error", err)
		return nil, err
	}
	dsMap, err := dbshare.DataSourcesMap(db)
	if err != nil {
		return nil, err
	}
	return &facetpg{db: db, dsm: dsMap}, nil
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
