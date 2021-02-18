package verifierpg

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/georgysavva/scany/sqlscan"
	"github.com/gnames/gnames/ent/verifier"
	mlib "github.com/gnames/gnlib/ent/matcher"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnparser"
	"github.com/gnames/gnparser/ent/parsed"
	log "github.com/sirupsen/logrus"
)

const (
	// resultsThreshold is the number of returned results for a match after
	// which we remove results with worst ParsingQuality. This step allows
	// to get rid of names of bacterial strains, 'sec.' names etc.
	resultsThreshold = 200
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
	ParseQuality        int
}

type matchSplit struct {
	noMatch   []*mlib.Match
	canonical []*mlib.Match
}

var namesQ = `
  SELECT canonical_id, name, data_source_id, record_id, name_string_id,
      local_id, outlink_id, accepted_record_id, accepted_name_id,
      accepted_name, classification, classification_ranks, parse_quality
    FROM verification where %s in (%s)`

// MatchRecords takes matches from gnmatcher and returns back data from
// the database that organizes data from database into matched records.
func (dgp verifierpg) MatchRecords(
	ctx context.Context,
	matches []mlib.Match,
) (map[string]*verifier.MatchRecord, error) {
	cfg := gnparser.NewConfig(gnparser.OptWithDetails(true))
	parser := gnparser.New(cfg)
	res := make(map[string]*verifier.MatchRecord)
	splitMatches := partitionMatches(matches)

	verifs, err := nameQuery(ctx, dgp.DB, splitMatches.canonical)
	if err != nil {
		log.Warnf("Cannot get matches data: %s", err)
		return res, err
	}
	res = dgp.produceResultData(splitMatches, parser, verifs)
	return res, nil
}

func (dgp verifierpg) produceResultData(
	ms matchSplit,
	parser gnparser.GNparser,
	v []*verif,
) map[string]*verifier.MatchRecord {

	// deal with NoMatch first
	mrs := make(map[string]*verifier.MatchRecord)
	for _, v := range ms.noMatch {
		mrs[v.ID] = &verifier.MatchRecord{
			InputID: v.ID,
			Input:   v.Name,
		}
	}

	verifMap := getVerifMap(v)
	for _, match := range ms.canonical {
		prsd := parser.ParseName(match.Name)
		if !prsd.Parsed {
			log.Fatalf("Cannot parse input '%s'. Should never happen at this point.", match.Name)
		}
		authors, year := processAuthorship(prsd.Authorship)

		mr := verifier.MatchRecord{
			InputID:         match.ID,
			Input:           match.Name,
			Cardinality:     int(prsd.Cardinality),
			CanonicalSimple: prsd.Canonical.Simple,
			CanonicalFull:   prsd.Canonical.Full,
			Authors:         authors,
			Year:            year,
			MatchType:       match.MatchType,
			Curation:        vlib.NotCurated,
		}
		for _, mi := range match.MatchItems {
			dgp.populateMatchRecord(mi, *match, &mr, parser, verifMap)
		}
		mrs[match.ID] = &mr
	}
	return mrs
}

func processAuthorship(au *parsed.Authorship) ([]string, int) {
	authors := make([]string, 0, 2)
	var year int
	if au == nil {
		return authors, year
	}

	authors = au.Authors

	year, err := strconv.Atoi(au.Year)
	if err == nil && !au.Original.Year.IsApproximate {
		return authors, year
	}

	if au.Combination != nil && au.Combination.Year != nil {
		year, _ = strconv.Atoi(au.Combination.Year.Value)
		if au.Combination.Year.IsApproximate {
			year = 0
		}
	}
	return authors, year
}

func (dgp *verifierpg) populateMatchRecord(
	mi mlib.MatchItem,
	m mlib.Match,
	mr *verifier.MatchRecord,
	parser gnparser.GNparser,
	verifMap map[string][]*verif,
) {
	verifRecs, ok := verifMap[mi.ID]
	if !ok {
		mr.MatchType = vlib.NoMatch
		return
	}

	sources := make(map[int]struct{})
	recsNum := len(verifRecs)
	var discardedExample string
	var discardedNum int
	for i, verifRec := range verifRecs {
		// all match types are the same, so we just take the first one to
		// expose it one level higher.
		if i == 0 {
			mr.MatchType = m.MatchType
		}

		// if there is a lot of records, most likely many of them are surrogates
		// that parser is not able to catch. Surrogates would parse with worst
		// parsing quality (4)
		if recsNum > resultsThreshold && verifRec.ParseQuality == 4 {
			if discardedExample == "" {
				discardedExample = verifRec.Name.String
			}
			discardedNum++
			continue
		}

		prsd := parser.ParseName(verifRec.Name.String)
		authors, year := processAuthorship(prsd.Authorship)

		currentRecordID := verifRec.RecordID.String
		currentName := verifRec.Name.String
		prsdCurrent := prsd
		currentCan := ""
		currentCanFull := ""
		if verifRec.AcceptedRecordID.Valid {
			currentRecordID = verifRec.AcceptedRecordID.String
			currentName = verifRec.AcceptedName.String
			prsdCurrent = parser.ParseName(currentName)
			if prsdCurrent.Parsed {
				currentCan = prsdCurrent.Canonical.Simple
				currentCanFull = prsdCurrent.Canonical.Full
			}
		}

		sources[verifRec.DataSourceID] = struct{}{}

		ds := dgp.DataSourcesMap[verifRec.DataSourceID]
		curation := ds.Curation

		if mr.Curation < curation {
			mr.Curation = curation
		}

		var dsID, matchCard, currCard, edDist, edDistStem int
		if m.MatchType != vlib.NoMatch {
			matchedCardinality := int(prsd.Cardinality)
			currentCardinality := int(prsdCurrent.Cardinality)

			dsID = verifRec.DataSourceID
			matchCard = matchedCardinality
			currCard = currentCardinality
			edDist = mi.EditDistance
			edDistStem = mi.EditDistanceStem
		}

		var outlink string
		if ds.OutlinkURL != "" && verifRec.OutlinkID.String != "" {
			outlink = strings.Replace(ds.OutlinkURL, "{}", verifRec.OutlinkID.String, 1)
		}

		resData := vlib.ResultData{
			RecordID:               verifRec.RecordID.String,
			LocalID:                verifRec.LocalID.String,
			Outlink:                outlink,
			DataSourceID:           dsID,
			DataSourceTitleShort:   ds.TitleShort,
			Curation:               curation,
			EntryDate:              ds.UpdatedAt,
			MatchedName:            verifRec.Name.String,
			MatchedCardinality:     matchCard,
			MatchedCanonicalSimple: prsd.Canonical.Simple,
			MatchedCanonicalFull:   prsd.Canonical.Full,
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
			ParsingQuality:         verifRec.ParseQuality,
		}

		mr.MatchResults = append(mr.MatchResults, &resData)

	}
	if discardedNum > 0 {
		log.Infof("Skipped %d low parsing quality names (e.g. '%s')", discardedNum,
			discardedExample)
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

func nameQuery(
	ctx context.Context,
	db *sql.DB,
	canMatches []*mlib.Match,
) ([]*verif, error) {

	var res []*verif
	if len(canMatches) == 0 {
		return res, nil
	}

	ids := getUUIDs(canMatches)
	idStr := "'" + strings.Join(ids, "','") + "'"
	q := fmt.Sprintf(namesQ, "canonical_id", idStr)
	err := sqlscan.Select(ctx, db, &res, q)
	if err != nil {
		return nil, err
	}
	return res, nil
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

func partitionMatches(matches []mlib.Match) matchSplit {
	ms := matchSplit{
		noMatch:   make([]*mlib.Match, 0, len(matches)),
		canonical: make([]*mlib.Match, 0, len(matches)),
	}
	for i := range matches {
		// TODO: handle v.VirusMatch case too.
		if matches[i].MatchType == vlib.NoMatch || matches[i].VirusMatch {
			ms.noMatch = append(ms.noMatch, &matches[i])
		} else {
			ms.canonical = append(ms.canonical, &matches[i])
		}
	}
	return ms
}
