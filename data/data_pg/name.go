package data_pg

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/sqlscan"
	"github.com/gnames/gnames/data"
	"github.com/gnames/gnames/domain/entity"
	gnm "github.com/gnames/gnmatcher/domain/entity"
	log "github.com/sirupsen/logrus"
	"gitlab.com/gogna/gnparser"
)

// MatchRecords connects result data to input name-string. Input name-string
// is a key.
type verif struct {
	CanonicalID         sql.NullString
	Canonical           sql.NullString
	CanonicalFull       sql.NullString
	Name                sql.NullString
	Cardinality         int
	RecordID            sql.NullString
	NameStringID        sql.NullString
	DataSourceID        int
	LocalID             sql.NullString
	OutlinkID           sql.NullString
	AcceptedRecordID    sql.NullString
	AcceptedNameID      sql.NullString
	AcceptedName        sql.NullString
	Classification      sql.NullString
	ClassificationRanks sql.NullString
}

type matchSplit struct {
	noMatch   []*gnm.Match
	canonical []*gnm.Match
}

var names_q = `
  SELECT canonical_id, name, data_source_id, record_id, name_string_id,
      local_id, outlink_id, accepted_record_id, accepted_name_id,
      accepted_name, classification, classification_ranks
    FROM verification where %s in (%s)`

// MatchRecords takes matches from gnmatcher and returns back data from
// the database that organizes data from database into matched records.
func (dgp DataGrabberPG) MatchRecords(matches []*gnm.Match) (map[string]*data.MatchRecord, error) {
	parser := gnparser.NewGNparser()
	res := make(map[string]*data.MatchRecord)
	splitMatches := partitionMatches(matches)

	verifs, err := nameQuery(dgp.DB, splitMatches.canonical)
	if err != nil {
		log.Warnf("Cannot get matches data: %s", err)
		return res, err
	}

	res = dgp.produceResultData(splitMatches, parser, verifs)
	return res, nil
}

func (dgp DataGrabberPG) produceResultData(
	ms matchSplit,
	parser gnparser.GNparser,
	v []*verif,
) map[string]*data.MatchRecord {

	// deal with NoMatch first
	mrs := make(map[string]*data.MatchRecord)
	for _, v := range ms.noMatch {
		mrs[v.ID] = &data.MatchRecord{
			InputID: v.ID,
			Input:   v.Name,
		}
	}

	verifMap := getVerifMap(v)
	for _, v := range ms.canonical {
		parsed := parser.ParseToObject(v.Name)
		if !parsed.Parsed {
			log.Fatalf("Cannot parse input '%s'. Should never happen at this point.", v.Name)
		}
		var authors []string
		if parsed.Authorship != nil {
			authors = parsed.Authorship.AllAuthors
		}
		mr := data.MatchRecord{
			InputID:         v.ID,
			Input:           v.Name,
			Cardinality:     int(parsed.Cardinality),
			CanonicalSimple: parsed.Canonical.GetSimple(),
			CanonicalFull:   parsed.Canonical.GetFull(),
			Authors:         authors,
			MatchType:       v.MatchType,
			CurationLevel:   entity.NotCurated,
		}
		for _, vv := range v.MatchItems {
			dgp.populateMatchRecord(vv, *v, &mr, parser, verifMap)
		}
		mrs[v.ID] = &mr
	}
	return mrs
}

func (dgp *DataGrabberPG) populateMatchRecord(
	mi gnm.MatchItem,
	m gnm.Match,
	mr *data.MatchRecord,
	parser gnparser.GNparser,
	verifMap map[string][]*verif,
) {
	v, ok := verifMap[mi.ID]
	if !ok {
		log.Fatalf("Cannot find verification records for %s.", mi.MatchStr)
	}

	sources := make(map[int]struct{})
	mr.MatchResults = make([]*entity.ResultData, len(v))
	for i, vv := range v {
		parsed := parser.ParseToObject(vv.Name.String)
		parsedCurrent := parser.ParseToObject(vv.AcceptedName.String)
		sources[vv.DataSourceID] = struct{}{}

		resData := entity.ResultData{
			ID:                     vv.RecordID.String,
			LocalID:                vv.LocalID.String,
			Outlink:                vv.OutlinkID.String,
			DataSourceID:           vv.DataSourceID,
			MatchedName:            vv.Name.String,
			MatchedCardinality:     vv.Cardinality,
			MatchedCanonicalSimple: parsed.Canonical.Simple,
			MatchedCanonicalFull:   parsed.Canonical.Full,
			CurrentName:            vv.AcceptedName.String,
			CurrentCardinality:     int(parsedCurrent.Cardinality),
			CurrentCanonicalSimple: parsedCurrent.Canonical.Simple,
			CurrentCanonicalFull:   parsedCurrent.Canonical.Full,
			IsSynonym:              vv.RecordID != vv.AcceptedRecordID,
			ClassificationPath:     vv.Classification.String,
			ClassificationRanks:    vv.ClassificationRanks.String,
			EditDistance:           mi.EditDistance,
			StemEditDistance:       mi.EditDistanceStem,
			MatchType:              m.MatchType,
		}
		mr.MatchResults[i] = &resData
		cl := dgp.DataSourcesMap[resData.DataSourceID].CurationLevel
		if mr.CurationLevel < cl {
			mr.CurationLevel = cl
		}
		if i == 0 {
			mr.MatchType = m.MatchType
		}
	}
	mr.DataSourcesNum = len(sources)
}

func getVerifMap(vs []*verif) map[string][]*verif {
	vm := make(map[string][]*verif)
	for _, v := range vs {
		vm[v.CanonicalID.String] = append(vm[v.CanonicalID.String], v)
	}
	return vm
}

func nameQuery(db *sql.DB, canMatches []*gnm.Match) ([]*verif, error) {
	ids := getUUIDs(canMatches)
	idStr := "'" + strings.Join(ids, "','") + "'"
	q := fmt.Sprintf(names_q, "canonical_id", idStr)
	var res []*verif
	ctx := context.Background()
	err := sqlscan.Select(ctx, db, &res, q)
	return res, err
}

func getUUIDs(matches []*gnm.Match) []string {
	set := make(map[string]struct{})
	for _, v := range matches {
		for _, vv := range v.MatchItems {
			if vv.EditDistance > 2 {
				continue
			}
			set[vv.ID] = struct{}{}
		}
	}
	res := make([]string, len(set))
	i := 0
	for k := range set {
		res[i] = k
		i++
	}
	return res
}

func partitionMatches(matches []*gnm.Match) matchSplit {
	ms := matchSplit{
		noMatch:   make([]*gnm.Match, 0, len(matches)),
		canonical: make([]*gnm.Match, 0, len(matches)),
	}
	for _, v := range matches {
		// TODO: handle v.VirusMatch case too.
		if v.MatchType == entity.NoMatch || v.VirusMatch {
			ms.noMatch = append(ms.noMatch, v)
		} else {
			ms.canonical = append(ms.canonical, v)
		}
	}
	return ms
}
