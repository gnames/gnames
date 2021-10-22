package verifierpg

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/georgysavva/scany/sqlscan"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gofrs/uuid"
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
  FROM data_sources order by id`

func (vf verifierpg) DataSources(ids ...int) ([]*vlib.DataSource, error) {
	q := dataSourcesQ
	if len(ids) > 0 {
		idsStrings := make([]string, len(ids))
		for i, v := range ids {
			idsStrings[i] = strconv.Itoa(v)
		}
		idsStr := strings.Join(idsStrings, ",")
		q = q + fmt.Sprintf(" WHERE id in (%s)", idsStr)
	}
	return vf.dataSourcesQuery(q)
}

func (vf verifierpg) dataSourcesQuery(q string) ([]*vlib.DataSource, error) {
	var dss []*dataSource
	ctx := context.Background()
	err := sqlscan.Select(ctx, vf.DB, &dss, q)
	res := make([]*vlib.DataSource, len(dss))
	for i, ds := range dss {
		dsItem := ds.convert()
		res[i] = &dsItem
	}
	return res, err
}
