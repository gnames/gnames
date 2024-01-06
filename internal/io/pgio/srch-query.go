package pgio

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gnames/gnames/pkg/ent/verif"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnparser"
	"github.com/gnames/gnparser/ent/parsed"
	"github.com/gnames/gnquery/ent/search"
	"github.com/gnames/gnuuid"
)

func setQuery(
	inp search.Input,
	spWordIDs []int,
	spWords string,
) (string, []interface{}) {
	spQ, args := spQuery(inp, spWordIDs, spWords)
	if inp.Author != "" {
		spQ, args = auQuery(spQ, inp, args)
	} else {
		spQ, args = noAuQuery(spQ, inp, args)
	}

	return spQ, args
}

func (p *pgio) runQuery(
	ctx context.Context,
	q string,
	args []interface{},
) (map[string]*verif.MatchRecord, error) {
	searches, err := p.searchQuery(ctx, q, args)
	if err != nil {
		return nil, err
	}
	return p.matchRecords(searches), nil
}

func (p *pgio) searchQuery(
	ctx context.Context,
	q string,
	args []interface{},
) ([]*verifSQL, error) {
	rows, err := p.db.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return rowsToVerifSQL(rows)
}

func (p *pgio) matchRecords(
	searches []*verifSQL,
) map[string]*verif.MatchRecord {
	pCfg := gnparser.NewConfig(gnparser.OptWithDetails(true))
	gnp := gnparser.New(pCfg)
	res := p.organizeByCanonicals(gnp, searches)
	return res
}

func (p *pgio) organizeByCanonicals(
	gnp gnparser.GNparser,
	searches []*verifSQL,
) map[string]*verif.MatchRecord {
	res := make(map[string]*verif.MatchRecord)
	for _, v := range searches {
		prsd := gnp.ParseName(v.Name.String)

		if !prsd.Parsed {
			slog.Error(
				"Should never happen",
				"error",
				fmt.Errorf("could not parse %s", v.Name.String),
			)
			continue // should never happen
		}
		if m, ok := res[prsd.Canonical.Full]; ok {
			mr := m.MatchResults
			mr = append(mr, p.matchRes(gnp, prsd, v))
			res[prsd.Canonical.Full].MatchResults = mr
		} else {
			mr := verif.MatchRecord{
				ID:              gnuuid.New(prsd.Canonical.Full).String(),
				Name:            prsd.Canonical.Full,
				Cardinality:     int(prsd.Cardinality),
				CanonicalSimple: prsd.Canonical.Simple,
				CanonicalFull:   prsd.Canonical.Full,
			}
			mr.MatchResults = []*vlib.ResultData{
				p.matchRes(gnp, prsd, v),
			}
			res[prsd.Canonical.Full] = &mr
		}
	}

	return res
}

func (p *pgio) matchRes(
	gnp gnparser.GNparser,
	prsd parsed.Parsed,
	v *verifSQL,
) *vlib.ResultData {
	authors, year := processAuthorship(prsd.Authorship)

	currentRecordID := v.RecordID.String
	currentName := v.Name.String
	prsdCurrent := prsd
	currentCan := ""
	currentCanFull := ""
	if v.AcceptedRecordID.Valid {
		currentRecordID = v.AcceptedRecordID.String
		currentName = v.AcceptedName.String
		prsdCurrent = gnp.ParseName(currentName)
		if prsdCurrent.Parsed {
			currentCan = prsdCurrent.Canonical.Simple
			currentCanFull = prsdCurrent.Canonical.Full
		}
	}
	matchedCardinality := int(prsd.Cardinality)
	currentCardinality := int(prsdCurrent.Cardinality)

	dsID := v.DataSourceID
	titleShort := p.dsm[dsID].TitleShort
	if titleShort == "" {
		titleShort = p.dsm[dsID].Title
	}

	var outlink string
	if p.dsm[dsID].OutlinkURL != "" && v.OutlinkID.String != "" {
		outlink = strings.Replace(
			p.dsm[dsID].OutlinkURL,
			"{}", v.OutlinkID.String, 1)
	}

	rd := vlib.ResultData{
		DataSourceID:           dsID,
		DataSourceTitleShort:   titleShort,
		Curation:               p.dsm[dsID].Curation,
		RecordID:               v.RecordID.String,
		LocalID:                v.LocalID.String,
		Outlink:                outlink,
		EntryDate:              p.dsm[dsID].UpdatedAt,
		ParsingQuality:         prsd.ParseQuality,
		MatchedName:            v.Name.String,
		MatchedCardinality:     matchedCardinality,
		MatchedAuthors:         authors,
		MatchedYear:            year,
		CurrentRecordID:        currentRecordID,
		CurrentName:            currentName,
		CurrentCardinality:     currentCardinality,
		CurrentCanonicalSimple: currentCan,
		CurrentCanonicalFull:   currentCanFull,
		IsSynonym:              v.RecordID != v.AcceptedRecordID,
		ClassificationPath:     v.Classification.String,
		ClassificationRanks:    v.ClassificationRanks.String,
		ClassificationIDs:      v.ClassificationIds.String,
		MatchType:              vlib.FacetedSearch,
	}
	return &rd
}

func queryEnd(
	q string,
	inp search.Input,
	args []interface{},
) (string, []interface{}) {
	if len(inp.DataSources) > 0 {
		args = append(args, inp.DataSources)
		q += fmt.Sprintf("\n    AND data_source_id = any($%d::int[])", len(args))
	}

	if inp.Year > 0 {
		args = append(args, inp.Year)
		q += fmt.Sprintf("\n    AND v.year = $%d", len(args))
	}

	if inp.YearRange != nil {
		if inp.YearStart > 0 {
			args = append(args, inp.YearStart)
			q += fmt.Sprintf("\n    AND v.year >= $%d", len(args))
		}

		if inp.YearEnd > 0 {
			args = append(args, inp.YearEnd)
			q += fmt.Sprintf("\n    AND v.year <= $%d", len(args))
		}
	}
	return q, args
}
