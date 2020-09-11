package data_pg

import (
	"context"
	"fmt"
	"time"

	"github.com/georgysavva/scany/sqlscan"
	"github.com/gnames/gnames/data"
	"github.com/gnames/gnames/domain/entity"
)

type dataSource struct {
	ID            int
	UUID          string
	Title         string
	TitleShort    string
	Version       string
	RevisionDate  string
	DOI           string
	Citation      string
	Authors       string
	Description   string
	WebsiteURL    string
	IsCurated     bool
	IsAutoCurated bool
	RecordCount   int
	UpdatedAt     time.Time
}

func (ds dataSource) convert() entity.DataSource {
	res := entity.DataSource{
		ID:           ds.ID,
		UUID:         ds.UUID,
		Title:        ds.Title,
		TitleShort:   ds.TitleShort,
		Version:      ds.Version,
		RevisionDate: ds.RevisionDate,
		DOI:          ds.DOI,
		Citation:     ds.Citation,
		Authors:      ds.Authors,
		Description:  ds.Description,
		RecordCount:  ds.RecordCount,
		UpdatedAt:    ds.UpdatedAt,
	}
	if ds.IsCurated {
		res.CurationLevel = entity.Curated
	} else if ds.IsAutoCurated {
		res.CurationLevel = entity.AutoCurated
	} else {
		res.CurationLevel = entity.NotCurated
	}
	return res
}

var data_sources_q = `
SELECT id, uuid, title, title_short, version, revision_date,
    doi, citation, authors, description, website_url,
    is_curated, is_auto_curated, record_count, updated_at
  FROM data_sources`

func (dgp DataGrabberPG) DataSources(id data.NullInt) ([]*entity.DataSource, error) {
	q := data_sources_q
	if id.Valid {
		q = q + fmt.Sprintf(" WHERE id = %d", id.Int)
	}
	return dgp.dataSourcesQuery(q)
}

func (dgp DataGrabberPG) dataSourcesQuery(q string) ([]*entity.DataSource, error) {
	var dss []*dataSource
	ctx := context.Background()
	err := sqlscan.Select(ctx, dgp.DB, &dss, q)
	res := make([]*entity.DataSource, len(dss))
	for i, ds := range dss {
		dsItem := ds.convert()
		res[i] = &dsItem
	}
	return res, err
}

