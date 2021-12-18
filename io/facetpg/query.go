package facetpg

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/georgysavva/scany/sqlscan"
	"github.com/gnames/gnames/ent/verifier"
	"github.com/gnames/gnames/io/internal/dbshare"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnparser"
	"github.com/gnames/gnparser/ent/parsed"
	"github.com/gnames/gnuuid"
	"github.com/lib/pq"
)

func (f *facetpg) setQuery() (string, []interface{}) {
	spQ, args := f.spQuery()
	if f.Author != "" {
		spQ, args = f.auQuery(spQ, args)
	} else {
		spQ, args = f.noAuQuery(spQ, args)
	}

	return spQ, args
}

func (f *facetpg) queryEnd(
	q string,
	args []interface{},
) (string, []interface{}) {
	if len(f.DataSourceIDs) > 0 {
		args = append(args, pq.Array(f.DataSourceIDs))
		q += fmt.Sprintf("\n    AND data_source_id = any($%d::int[])", len(args))
	}

	if f.Year > 0 {
		args = append(args, f.Year)
		q += fmt.Sprintf("\n    AND v.year = $%d", len(args))
	}

	if f.YearRange != nil {
		if f.YearStart > 0 {
			args = append(args, f.YearStart)
			q += fmt.Sprintf("\n    AND v.year >= $%d", len(args))
		}

		if f.YearEnd > 0 {
			args = append(args, f.YearEnd)
			q += fmt.Sprintf("\n    AND v.year <= $%d", len(args))
		}
	}
	return q, args
}

func (f *facetpg) runQuery(
	ctx context.Context,
	q string,
	args []interface{},
) (map[string]*verifier.MatchRecord, error) {
	searches, err := f.searchQuery(ctx, q, args)
	if err != nil {
		return nil, err
	}
	return f.matchRecords(searches), nil
}

func (f *facetpg) searchQuery(
	ctx context.Context,
	q string,
	args []interface{},
) ([]*dbshare.VerifSQL, error) {
	var res []*dbshare.VerifSQL
	err := sqlscan.Select(ctx, f.db, &res, q, args...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (f *facetpg) matchRecords(
	searches []*dbshare.VerifSQL,
) map[string]*verifier.MatchRecord {

	pCfg := gnparser.NewConfig(gnparser.OptWithDetails(true))
	gnp := gnparser.New(pCfg)
	res := f.organizeByCanonicals(gnp, searches)
	return res
}

func (f *facetpg) organizeByCanonicals(
	gnp gnparser.GNparser,
	searches []*dbshare.VerifSQL,
) map[string]*verifier.MatchRecord {
	res := make(map[string]*verifier.MatchRecord)
	for _, v := range searches {
		prsd := gnp.ParseName(v.Name.String)

		if !prsd.Parsed {
			log.Printf("Could not parse %s", v.Name.String)
			continue // should never happen
		}
		if m, ok := res[prsd.Canonical.Full]; ok {
			mr := m.MatchResults
			mr = append(mr, f.matchRes(gnp, prsd, v))
			res[prsd.Canonical.Full].MatchResults = mr
		} else {
			mr := verifier.MatchRecord{
				ID:              gnuuid.New(prsd.Canonical.Full).String(),
				Name:            prsd.Canonical.Full,
				Cardinality:     int(prsd.Cardinality),
				CanonicalSimple: prsd.Canonical.Simple,
				CanonicalFull:   prsd.Canonical.Full,
				MatchType:       vlib.FacetedSearch,
				Curation:        vlib.NotCurated,
			}
			mr.MatchResults = []*vlib.ResultData{
				f.matchRes(gnp, prsd, v),
			}
			res[prsd.Canonical.Full] = &mr
		}
	}

	return res
}

func (f *facetpg) matchRes(
	gnp gnparser.GNparser,
	prsd parsed.Parsed,
	v *dbshare.VerifSQL,
) *vlib.ResultData {
	authors, year := dbshare.ProcessAuthorship(prsd.Authorship)

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
	titleShort := f.dsm[dsID].TitleShort
	if titleShort == "" {
		titleShort = f.dsm[dsID].Title
	}

	var outlink string
	if f.dsm[dsID].OutlinkURL != "" && v.OutlinkID.String != "" {
		outlink = strings.Replace(
			f.dsm[dsID].OutlinkURL,
			"{}", v.OutlinkID.String, 1)
	}

	rd := vlib.ResultData{
		DataSourceID:           dsID,
		DataSourceTitleShort:   titleShort,
		Curation:               f.dsm[dsID].Curation,
		RecordID:               v.RecordID.String,
		LocalID:                v.LocalID.String,
		Outlink:                outlink,
		EntryDate:              f.dsm[dsID].UpdatedAt,
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
