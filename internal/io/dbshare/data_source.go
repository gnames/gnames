package dbshare

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/georgysavva/scany/sqlscan"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type dataSource struct {
	ID             int
	UUID           string
	Title          string
	TitleShort     string
	Version        string
	RevisionDate   string
	DOI            string
	Citation       string
	Authors        string
	Description    string
	WebsiteURL     string
	OutlinkURL     string
	IsOutlinkReady bool
	IsCurated      bool
	IsAutoCurated  bool
	RecordCount    int
	UpdatedAt      time.Time
}

func DataSourcesMap(db *sql.DB) (map[int]*vlib.DataSource, error) {
	res := make(map[int]*vlib.DataSource)
	dss, err := DataSources(db)
	if err != nil {
		slog.Error("Cannot init DataSources data", "error", err)
		return res, err
	}
	for _, ds := range dss {
		res[ds.ID] = ds
	}
	return res, nil
}

func (ds dataSource) convert() vlib.DataSource {
	res := vlib.DataSource{
		ID:             ds.ID,
		UUID:           "",
		Title:          ds.Title,
		TitleShort:     ds.TitleShort,
		Version:        ds.Version,
		RevisionDate:   ds.RevisionDate,
		DOI:            ds.DOI,
		Citation:       ds.Citation,
		Authors:        ds.Authors,
		WebsiteURL:     ds.WebsiteURL,
		OutlinkURL:     ds.OutlinkURL,
		IsOutlinkReady: ds.IsOutlinkReady,
		Description:    ds.Description,
		RecordCount:    ds.RecordCount,
		UpdatedAt:      ds.UpdatedAt.Format("2006-01-02"),
	}
	if ds.UUID != uuid.Nil.String() {
		res.UUID = ds.UUID
	}
	if ds.IsCurated {
		res.Curation = vlib.Curated
	} else if ds.IsAutoCurated {
		res.Curation = vlib.AutoCurated
	} else {
		res.Curation = vlib.NotCurated
	}
	return res
}

var dataSourcesQ = `
SELECT id, uuid, title, title_short, version, revision_date,
    doi, citation, authors, description, website_url, outlink_url,
    is_outlink_ready, is_curated, is_auto_curated, record_count, updated_at
  FROM data_sources`

func DataSources(db *sql.DB, ids ...int) ([]*vlib.DataSource, error) {
	q := dataSourcesQ
	if len(ids) > 0 {
		q += " WHERE id = any($1)"
	}
	q += " order by id"
	return dataSourcesQuery(db, q, ids)
}

func dataSourcesQuery(db *sql.DB, q string, ids []int) ([]*vlib.DataSource, error) {
	var dss []*dataSource
	var err error
	ctx := context.Background()
	if len(ids) > 0 {
		err = sqlscan.Select(ctx, db, &dss, q, pq.Array(ids))
	} else {
		err = sqlscan.Select(ctx, db, &dss, q)
	}
	res := make([]*vlib.DataSource, len(dss))
	for i, ds := range dss {
		dsItem := ds.convert()
		res[i] = &dsItem
	}
	return res, err
}
