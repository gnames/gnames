package data_pg

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/georgysavva/scany/sqlscan"
	"github.com/gnames/gnames/data"
	mlib "github.com/gnames/gnlib/domain/entity/matcher"
	vlib "github.com/gnames/gnlib/domain/entity/verifier"
	log "github.com/sirupsen/logrus"
	"gitlab.com/gogna/gnparser"
	"gitlab.com/gogna/gnparser/pb"
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
	noMatch   []*mlib.Match
	canonical []*mlib.Match
}

var names_q = `
  SELECT canonical_id, name, data_source_id, record_id, name_string_id,
      local_id, outlink_id, accepted_record_id, accepted_name_id,
      accepted_name, classification, classification_ranks
    FROM verification where %s in (%s)`

// MatchRecords takes matches from gnmatcher and returns back data from
// the database that organizes data from database into matched records.
func (dgp DataGrabberPG) MatchRecords(matches []*mlib.Match) (map[string]*data.MatchRecord, error) {
	for _, v := range matches {
		if v.Name == "Acacia vestita may" {
			log.Debugf("ACACIA: %+v", v)
		}
	}
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
	for _, match := range ms.canonical {
		parsed := parser.ParseToObject(match.Name)
		if !parsed.Parsed {
			log.Fatalf("Cannot parse input '%s'. Should never happen at this point.", match.Name)
		}
		authors, year := processAuthorship(parsed.Authorship)

		mr := data.MatchRecord{
			InputID:         match.ID,
			Input:           match.Name,
			Cardinality:     int(parsed.Cardinality),
			CanonicalSimple: parsed.Canonical.GetSimple(),
			CanonicalFull:   parsed.Canonical.GetFull(),
			Authors:         authors,
			Year:            year,
			MatchType:       match.MatchType,
			CurationLevel:   vlib.NotCurated,
		}
		for _, mi := range match.MatchItems {
			dgp.populateMatchRecord(mi, *match, &mr, parser, verifMap)
		}
		mrs[match.ID] = &mr
	}
	return mrs
}

func processAuthorship(au *pb.Authorship) ([]string, int) {
	authors := make([]string, 0, 2)
	var year int
	if au == nil {
		return authors, year
	}

	authors = au.AllAuthors

	if au.Original != nil && au.Original.Year != "" {
		yr, err := strconv.Atoi(au.Original.Year)
		if err == nil && !au.Original.ApproximateYear {
			year = yr
		}
	}
	if au.Combination != nil && au.Combination.Year != "" {
		if year > 0 {
			return authors, 0
		}

		yr, err := strconv.Atoi(au.Original.Year)
		if err == nil && !au.Original.ApproximateYear {
			year = yr
		}
	}
	return authors, year
}

func (dgp *DataGrabberPG) populateMatchRecord(
	mi mlib.MatchItem,
	m mlib.Match,
	mr *data.MatchRecord,
	parser gnparser.GNparser,
	verifMap map[string][]*verif,
) {
	verifRecs, ok := verifMap[mi.ID]
	if !ok {
		mr.MatchType = vlib.NoMatch
		return
	}

	sources := make(map[int]struct{})
	for i, verifRec := range verifRecs {
		parsed := parser.ParseToObject(verifRec.Name.String)
		authors, year := processAuthorship(parsed.Authorship)

		currentRecordID := verifRec.RecordID.String
		currentName := verifRec.Name.String
		parsedCurrent := parsed
		currentCan := ""
		currentCanFull := ""
		if verifRec.AcceptedRecordID.Valid {
			currentRecordID = verifRec.AcceptedRecordID.String
			currentName = verifRec.AcceptedName.String
			parsedCurrent = parser.ParseToObject(currentName)
			if parsedCurrent.Parsed {
				currentCan = parsedCurrent.Canonical.Simple
				currentCanFull = parsedCurrent.Canonical.Full
			}
		}

		sources[verifRec.DataSourceID] = struct{}{}

		ds := dgp.DataSourcesMap[verifRec.DataSourceID]
		curationLevel := ds.CurationLevel

		if mr.CurationLevel < curationLevel {
			mr.CurationLevel = curationLevel
		}

		var dsID, matchCard, currCard, edDist, edDistStem *int
		if m.MatchType != vlib.NoMatch {
			matchedCardinality := int(parsed.Cardinality)
			currentCardinality := int(parsedCurrent.Cardinality)

			dsID = &verifRec.DataSourceID
			matchCard = &matchedCardinality
			currCard = &currentCardinality
			edDist = &mi.EditDistance
			edDistStem = &mi.EditDistanceStem
		}

		resData := vlib.ResultData{
			ID:                     verifRec.RecordID.String,
			LocalID:                verifRec.LocalID.String,
			Outlink:                verifRec.OutlinkID.String,
			DataSourceID:           dsID,
			DataSrouceTitleShort:   ds.TitleShort,
			CurationLevel:          curationLevel,
			CurationLevelString:    curationLevel.String(),
			EntryDate:              ds.UpdatedAt.Format("2006-01-02"),
			MatchedName:            verifRec.Name.String,
			MatchedCardinality:     matchCard,
			MatchedCanonicalSimple: parsed.Canonical.Simple,
			MatchedCanonicalFull:   parsed.Canonical.Full,
			MatchedAuthors:         authors,
			MatchedYear:            year,
			CurrentRecordID:        currentRecordID,
			CurrentName:            currentName,
			CurrentCardinality:     currCard,
			CurrentCanonicalSimple: currentCan,
			CurrentCanonicalFull:   currentCanFull,
			IsSynonym:              verifRec.RecordID != verifRec.AcceptedRecordID,
			ClassificationPath:     verifRec.Classification.String,
			ClassificationRanks:    verifRec.ClassificationRanks.String,
			EditDistance:           edDist,
			StemEditDistance:       edDistStem,
			MatchType:              m.MatchType,
		}
		mr.MatchResults = append(mr.MatchResults, &resData)
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

func nameQuery(db *sql.DB, canMatches []*mlib.Match) ([]*verif, error) {
	var res []*verif
	if len(canMatches) == 0 {
		return res, nil
	}

	ids := getUUIDs(canMatches)
	idStr := "'" + strings.Join(ids, "','") + "'"
	q := fmt.Sprintf(names_q, "canonical_id", idStr)
	ctx := context.Background()
	err := sqlscan.Select(ctx, db, &res, q)
	return res, err
}

func getUUIDs(matches []*mlib.Match) []string {
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

func partitionMatches(matches []*mlib.Match) matchSplit {
	ms := matchSplit{
		noMatch:   make([]*mlib.Match, 0, len(matches)),
		canonical: make([]*mlib.Match, 0, len(matches)),
	}
	for _, v := range matches {
		// TODO: handle v.VirusMatch case too.
		if v.MatchType == vlib.NoMatch || v.VirusMatch {
			ms.noMatch = append(ms.noMatch, v)
		} else {
			ms.canonical = append(ms.canonical, v)
		}
	}
	return ms
}
