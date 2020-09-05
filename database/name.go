package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/sqlscan"
	"github.com/gnames/gnames/model"
	gnm "github.com/gnames/gnmatcher/model"
	log "github.com/sirupsen/logrus"
	"gitlab.com/gogna/gnparser"
)

// MatchRecords connects result data to input name-string. Input name-string
// is a key.

type InputUUID string

type MatchRecord struct {
	InputID        string
	Input          string
	Score          int
	MatchType      model.MatchType
	CurationLevel  model.CurationLevel
	DataSourcesNum int
	ResultData     []*model.ResultData
}

type verif struct {
	CanonicalID         sql.NullString
	CanonicalFullID     sql.NullString
	Name                sql.NullString
	Cardinality         int
	DataSourceID        int
	RecordID            sql.NullString
	NameStringID        sql.NullString
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
  SELECT canonical_id, canonical_full_id, name,
      cardinality, data_source_id, record_id, name_string_id,
      local_id, outlink_id, accepted_record_id, accepted_name_id,
      accepted_name, classification, classification_ranks
    FROM verification where %s in (%s)`

// MatchRecords takes matches from gnmatcher and returns back data from
// the database that organizes data from database into matched records.
func MatchRecords(
	db *sql.DB,
	dss map[int]*model.DataSource,
	matches []*gnm.Match,
) (map[InputUUID]MatchRecord, error) {
	parser := gnparser.NewGNparser()
	res := make(map[InputUUID]MatchRecord)
	splitMatches := partitionMatches(matches)

	verifs, err := nameQuery(db, splitMatches.canonical)
	if err != nil {
		return res, err
	}

	res = produceResultData(splitMatches, dss, parser, verifs)
	return res, nil
}

func produceResultData(
	ms matchSplit,
	dss map[int]*model.DataSource,
	parser gnparser.GNparser,
	v []*verif,
) map[InputUUID]MatchRecord {

	// deal with NoMatch first
	mrs := make(map[InputUUID]MatchRecord)
	for _, v := range ms.noMatch {
		mrs[InputUUID(v.ID)] = MatchRecord{
			InputID: v.ID,
			Input:   v.Name,
		}
	}

	verifMap := getVerifMap(v)
	for _, v := range ms.canonical {
		mr := MatchRecord{
			InputID:       v.ID,
			Input:         v.Name,
			MatchType:     v.MatchType,
			CurationLevel: model.NotCurated,
		}
		for _, vv := range v.MatchItems {
			mr.populateMatchRecord(vv, *v, dss, parser, verifMap)
		}
		mr.DataSourcesNum = len(dss)
		mrs[InputUUID(v.ID)] = mr
	}
	return mrs
}

func (mr *MatchRecord) populateMatchRecord(
	mi gnm.MatchItem,
	m gnm.Match,
	dss map[int]*model.DataSource,
	parser gnparser.GNparser,
	verifMap map[string][]*verif,
) {
	v, ok := verifMap[mi.ID]
	if !ok {
		log.Fatalf("Cannot find verification records for %s.", mi.MatchStr)
	}

	sources := make(map[int]struct{})
	mr.ResultData = make([]*model.ResultData, len(v))
	for i, vv := range v {
		parsed := parser.ParseToObject(vv.Name.String)
		parsedCurrent := parser.ParseToObject(vv.AcceptedName.String)
		sources[vv.DataSourceID] = struct{}{}

		resData := model.ResultData{
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
		mr.ResultData[i] = &resData
		cl := dss[resData.DataSourceID].CurationLevel
		if mr.CurationLevel < cl {
			mr.CurationLevel = cl
		}
		if i == 0 {
			mr.MatchType = m.MatchType
			log.Println(mr.MatchType)
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
		if v.MatchType == model.NoMatch || v.VirusMatch {
			ms.noMatch = append(ms.noMatch, v)
		} else {
			ms.canonical = append(ms.canonical, v)
		}
	}
	return ms
}
