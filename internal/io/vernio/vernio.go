package vernio

import (
	"context"
	"fmt"

	"github.com/gnames/gnames/pkg/config"
	"github.com/gnames/gnames/pkg/ent/pg"
	"github.com/gnames/gnames/pkg/ent/vern"
	"github.com/gnames/gnlib/ent/verifier"
	vlib "github.com/gnames/gnlib/ent/verifier"
)

type vernio struct {
	cfg config.Config
	db  pg.PG
}

func New(cfg config.Config, db pg.PG) vern.Vernaculars {
	res := vernio{
		cfg: cfg,
		db:  db,
	}
	return &res
}

func (v *vernio) AddVernacularNames(langs []string, names []vlib.Name) ([]vlib.Name, error) {
	ctx := context.Background()
	recordsMap := vernacularRecords(names)
	// did not find any records to search for vernaculars
	if len(recordsMap) == 0 {
		return names, nil
	}

	records := make([]vern.Record, len(recordsMap))
	var count int
	for k := range recordsMap {
		records[count] = k
		count++
	}

	verns, err := v.db.GetVernaculars(ctx, records, langs)
	if err != nil {
		return names, fmt.Errorf("vernio.AddVernacularNames: failed to get vernaculars: %w", err)
	}

	for k, v := range verns {
		for i := range recordsMap[k] {
			recordsMap[k][i].Vernaculars = v
		}
	}

	return names, nil
}

func updateNames(names []vlib.Name, verns []vlib.Vernacular) []vlib.Name {
	return names
}

func vernacularRecords(names []vlib.Name) map[vern.Record][]*verifier.ResultData {
	res := make(map[vern.Record][]*verifier.ResultData)
	for _, v := range names {
		for _, result := range v.Results {
			if result.ClassificationPath != "" && result.CurrentName != "" {
				rec := vern.Record{
					DataSourceID: result.DataSourceID,
					RecordID:     result.RecordID,
				}
				res[rec] = append(res[rec], result)
			}
		}
		if v.BestResult != nil && v.BestResult.ClassificationPath != "" && v.BestResult.CurrentName != "" {
			rec := vern.Record{
				DataSourceID: v.BestResult.DataSourceID,
				RecordID:     v.BestResult.RecordID,
			}
			res[rec] = append(res[rec], v.BestResult)
		}
	}
	return res
}
