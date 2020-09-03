package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/georgysavva/scany/sqlscan"
	"github.com/gnames/gnames/model"
)

type dataSource struct {
	ID            int
	UUID          string
	Title         string
	TitleShort    string
	Version       string
	CreationDate  string
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

func (ds dataSource) convert() model.DataSource {
	res := model.DataSource{
		ID:           ds.ID,
		UUID:         ds.UUID,
		Title:        ds.Title,
		TitleShort:   ds.TitleShort,
		Version:      ds.Version,
		CreationDate: ds.CreationDate,
		DOI:          ds.DOI,
		Citation:     ds.Citation,
		Authors:      ds.Authors,
		Description:  ds.Description,
		RecordCount:  ds.RecordCount,
		UpdatedAt:    ds.UpdatedAt,
	}
	if ds.IsCurated {
		res.CurationLevel = model.Curated
	} else if ds.IsAutoCurated {
		res.CurationLevel = model.AutoCurated
	} else {
		res.CurationLevel = model.NotCurated
	}
	return res
}

var data_sources_q = `
  SELECT id, uuid, title, title_short, version, creation_date,
    doi, citation, authors, description, website_url,
    is_curated, is_auto_curated, record_count, updated_at
  FROM data_sources`

func GetDataSources(db *sql.DB) ([]*model.DataSource, error) {
	return dataSourcesQuery(db, data_sources_q)
}

func GetDataSource(db *sql.DB, id int) ([]*model.DataSource, error) {
	q := data_sources_q + fmt.Sprintf(" where id = %d", id)
	return dataSourcesQuery(db, q)
}

func dataSourcesQuery(db *sql.DB, q string) ([]*model.DataSource, error) {
	var dss []*dataSource
	ctx := context.Background()
	err := sqlscan.Select(ctx, db, &dss, q)
	res := make([]*model.DataSource, len(dss))
	for i, ds := range dss {
		dsItem := ds.convert()
		res[i] = &dsItem
	}
	return res, err
}
