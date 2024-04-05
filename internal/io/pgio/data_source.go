package pgio

import (
	"context"
	"fmt"
	"time"

	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (p *pgio) dataSources(ids ...int) ([]*vlib.DataSource, error) {
	var err error
	var rows pgx.Rows
	var dss []*dataSource
	ctx := context.Background()

	idsWere := "WHERE id = any($1)"
	q := `
SELECT id, uuid, title, title_short, version, revision_date,
    doi, citation, authors, description, website_url, outlink_url,
    is_outlink_ready, is_curated, has_taxon_data, is_auto_curated,
	  record_count, updated_at
	FROM data_sources
	%s
	ORDER BY id
  `
	insert := ""
	if len(ids) > 0 {
		insert = idsWere
	}
	q = fmt.Sprintf(q, insert)

	if len(ids) == 0 {
		rows, err = p.db.Query(ctx, q)
	} else {
		rows, err = p.db.Query(ctx, q, ids)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ds dataSource
		err = rows.Scan(
			&ds.ID, &ds.UUID, &ds.Title, &ds.TitleShort, &ds.Version,
			&ds.RevisionDate, &ds.DOI, &ds.Citation, &ds.Authors,
			&ds.Description, &ds.WebsiteURL, &ds.OutlinkURL,
			&ds.IsOutlinkReady, &ds.IsCurated, &ds.HasTaxonData,
			&ds.IsAutoCurated, &ds.RecordCount, &ds.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		dss = append(dss, &ds)
	}

	res := make([]*vlib.DataSource, len(dss))

	for i := range dss {
		dsItem := dss[i].convert()
		res[i] = &dsItem
	}
	return res, err
}

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
	HasTaxonData   bool
	IsAutoCurated  bool
	RecordCount    int
	UpdatedAt      time.Time
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
		HasTaxonData:   ds.HasTaxonData,
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
