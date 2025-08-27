package pgio

import (
	"context"
	"fmt"

	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnames/pkg/ent/pg"
	"github.com/gnames/gnames/pkg/ent/verif"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnparser"
	"github.com/gnames/gnquery/ent/search"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgio struct {
	db      *pgxpool.Pool
	dsm     map[int]*vlib.DataSource
	gnpPool chan gnparser.GNparser
}

func New(cfg config.Config) (pg.PG, error) {
	res, err := conn(cfg)
	if err != nil {
		return nil, fmt.Errorf("new PG instance failed: %w", err)
	}

	dsm, err := res.dataSourcesMap()
	if err != nil {
		return nil, fmt.Errorf("could not get data sources map: %w", err)
	}
	res.dsm = dsm

	poolSize := 5
	gnpPool := make(chan gnparser.GNparser, poolSize)
	for range poolSize {
		cfgGNP := gnparser.NewConfig(gnparser.OptWithDetails(true))
		gnpPool <- gnparser.New(cfgGNP)
	}
	res.gnpPool = gnpPool

	return res, nil
}

func (p *pgio) DataSourcesMap() map[int]*vlib.DataSource {
	return p.dsm
}

func (p *pgio) dataSourcesMap() (map[int]*vlib.DataSource, error) {
	res := make(map[int]*vlib.DataSource)
	dss, err := p.dataSources()
	if err != nil {
		return res, fmt.Errorf("pgio.dataSourceMap: %w", err)
	}
	for _, ds := range dss {
		res[ds.ID] = ds
	}
	return res, nil
}

func (p *pgio) MatchRecordsMap(
	ctx context.Context,
	splitMatches verif.MatchSplit,
	input vlib.Input,
) (map[string]*verif.MatchRecord, error) {

	var err error
	res := make(map[string]*verif.MatchRecord)
	cfg := gnparser.NewConfig(gnparser.OptWithDetails(true))
	parser := gnparser.New(cfg)

	// find matches for canonicals
	verCan, err := p.nameQuery(ctx, splitMatches.Canonical, input)
	if err != nil {
		return res, fmt.Errorf("matching results to canonical failed: %w", err)
	}

	// find matches for viruses
	verVir, err := p.virusQuery(ctx, splitMatches.Virus, input)
	if err != nil {
		return res, fmt.Errorf("getting virus data failed: %w", err)
	}

	// convert matches to intermediate results
	res = p.produceResultData(splitMatches, parser, verCan, verVir)

	return res, nil

}

// NameByID finds a name-string in the database by its ID.
// It returns all matches for the name-string accoring to
// NameStringInput settings. It can limit results to the best match only,
// it can also filter results by data-sources.
func (p *pgio) NameByID(inp vlib.NameStringInput) (*verif.MatchRecord, error) {
	ctx := context.Background()
	q, args := idQuery(inp)
	vSQL, err := p.idQueryRun(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("getting name by ID failed: %w", err)
	}

	return p.idData(vSQL), nil
}

func (p *pgio) NameStringByID(id string) (string, error) {
	ctx := context.Background()

	var res string
	row :=
		p.db.QueryRow(ctx, "SELECT name FROM name_strings WHERE id = $1", id)
	err := row.Scan(&res)
	if err != nil {
		return res, fmt.Errorf("pgio.NameStringByID: %w", err)
	}

	return res, nil
}

func (p *pgio) SearchRecordsMap(
	ctx context.Context,
	input search.Input,
	spWordIDs []int,
	spWord string,
) (map[string]*verif.MatchRecord, error) {
	q, args := setQuery(input, spWordIDs, spWord)
	res, err := p.runQuery(ctx, q, args)
	if err != nil {
		return nil, fmt.Errorf("pgio.SearchRecordsMap: %w", err)
	}
	return res, nil
}
