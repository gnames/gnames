package verifierpg

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/sqlscan"
	"github.com/gnames/gnames/ent/verifier"
	"github.com/gnames/gnames/io/internal/dbshare"
	mlib "github.com/gnames/gnlib/ent/matcher"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnparser"
	"github.com/lib/pq"
	"github.com/rs/zerolog/log"
)

const (
	// resultsThreshold is the number of returned results for a match after
	// which we remove results with worst ParsingQuality. This step allows
	// to get rid of names of bacterial strains, 'sec.' names etc.
	resultsThreshold = 100
)

type matchSplit struct {
	noMatch   []*mlib.Match
	virus     []*mlib.Match
	canonical []*mlib.Match
}

var namesQ = fmt.Sprintf(`
SELECT %s
FROM verification v
WHERE canonical_id = ANY($1::uuid[])
`, dbshare.QueryFields)

var virusQ = fmt.Sprintf(`
SELECT %s
FROM verification v
  WHERE name_string_id = ANY($1::uuid[])
`, dbshare.QueryFields)

// MatchRecords takes matches from gnmatcher and returns back data from
// the database that organizes data from database into matched records.
func (dgp verifierpg) MatchRecords(
	ctx context.Context,
	matches []mlib.Match,
	input vlib.Input,
) (map[string]*verifier.MatchRecord, error) {
	cfg := gnparser.NewConfig(gnparser.OptWithDetails(true))
	parser := gnparser.New(cfg)
	var res map[string]*verifier.MatchRecord

	// separate NoMatch, Virus, and matches
	splitMatches := partitionMatches(matches)

  fmt.Printf("MATCHES: %#v\n\n", matches)

	// find matches for canonicals
	verCan, err := dgp.nameQuery(ctx, splitMatches.canonical, input)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot get matches data")
		return res, err
	}

	// find matches for viruses
	verVir, err := dgp.virusQuery(ctx, splitMatches.virus, input)
	if err != nil {
		log.Warn().Err(err).Msg("Cannot get virus data")
		return res, err
	}

	// convert matches to intermediate results
	res = dgp.produceResultData(splitMatches, parser, verCan, verVir)

	return res, nil
}

func (dgp verifierpg) produceResultData(
	ms matchSplit,
	parser gnparser.GNparser,
	verCan []*dbshare.VerifSQL,
	verVir []*dbshare.VerifSQL,
) map[string]*verifier.MatchRecord {

	// deal with NoMatch first
	allMatchRecs := make(map[string]*verifier.MatchRecord)
	for _, v := range ms.noMatch {
		allMatchRecs[v.ID] = &verifier.MatchRecord{
			ID:   v.ID,
			Name: v.Name,
		}
	}

	vers := verCan
	vers = append(vers, verVir...)
	// organize results by either CanonicalID or
	// NameID (for viruses)
	verifMap := getVerifMap(vers)

	// deal with Viruses
	for _, match := range ms.virus {
		mr := verifier.MatchRecord{
			ID:       match.ID,
			Name:     match.Name,
			Overload: len(match.MatchItems) > 20,
		}
		for _, mi := range match.MatchItems {
			dgp.populateVirusMatchRecord(mi, *match, &mr, verifMap)
		}
		allMatchRecs[match.ID] = &mr
	}

	// deal with Canonicals
	for _, match := range ms.canonical {
		// TODO check if parsing affects speed too much
		prsd := parser.ParseName(match.Name)

		if !prsd.Parsed {
			log.Fatal().Err(errors.New("cannot parse")).Str("name", match.Name).
				Msg("Should never happen")
		}
		authors, year := dbshare.ProcessAuthorship(prsd.Authorship)

		mr := verifier.MatchRecord{
			ID:              match.ID,
			Name:            match.Name,
			Cardinality:     int(prsd.Cardinality),
			CanonicalSimple: prsd.Canonical.Simple,
			CanonicalFull:   prsd.Canonical.Full,
			Authors:         authors,
			Year:            year,
		}

		for _, mi := range match.MatchItems {
			dgp.populateMatchRecord(mi, *match, &mr, parser, verifMap)
		}
		allMatchRecs[match.ID] = &mr
	}
	return allMatchRecs
}

func (dgp *verifierpg) populateVirusMatchRecord(
	mi mlib.MatchItem,
	m mlib.Match,
	mr *verifier.MatchRecord,
	verifMap map[string][]*dbshare.VerifSQL,
) {
	verifRecs, ok := verifMap[mi.ID]
	if !ok {
		log.Fatal().Err(fmt.Errorf("no match for %s", mi.ID))
	}
	sources := make(map[int]struct{})
	for _, verifRec := range verifRecs {
		sources[verifRec.DataSourceID] = struct{}{}

		ds := dgp.dataSourcesMap[verifRec.DataSourceID]
		curation := ds.Curation

		var outlink string
		if ds.OutlinkURL != "" && verifRec.OutlinkID.String != "" {
			outlink = strings.Replace(ds.OutlinkURL, "{}", verifRec.OutlinkID.String, 1)
		}

		titleShort := ds.TitleShort
		if titleShort == "" {
			titleShort = ds.Title
		}

		resData := vlib.ResultData{
			RecordID:             verifRec.RecordID.String,
			LocalID:              verifRec.LocalID.String,
			Outlink:              outlink,
			DataSourceID:         verifRec.DataSourceID,
			DataSourceTitleShort: titleShort,
			Curation:             curation,
			EntryDate:            ds.UpdatedAt,
			MatchedName:          verifRec.Name.String,
			ClassificationPath:   verifRec.Classification.String,
			ClassificationRanks:  verifRec.ClassificationRanks.String,
			ClassificationIDs:    verifRec.ClassificationIds.String,
			MatchType:            m.MatchType,
		}

		mr.MatchResults = append(mr.MatchResults, &resData)
	}
	mr.DataSourcesNum = len(sources)
}

func (dgp *verifierpg) populateMatchRecord(
	mItm mlib.MatchItem,
	m mlib.Match,
	mRec *verifier.MatchRecord,
	parser gnparser.GNparser,
	verifMap map[string][]*dbshare.VerifSQL,
) {
	verifRecs, ok := verifMap[mItm.ID]
	if !ok {
		return
	}

	sources := make(map[int]struct{})
	recsNum := len(verifRecs)
	var discardedExample string
	var discardedNum int
	for _, verifRec := range verifRecs {
		// if there is a lot of records, most likely many of them are surrogates
		// that parser is not able to catch. Surrogates would parse with worst
		// parsing quality (4)
		mRec.Overload = recsNum > resultsThreshold
		if recsNum > resultsThreshold && verifRec.ParseQuality == 4 {
			if discardedExample == "" {
				discardedExample = verifRec.Name.String
			}
			discardedNum++
			continue
		}

		prsd := parser.ParseName(verifRec.Name.String)
		authors, year := dbshare.ProcessAuthorship(prsd.Authorship)

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

		ds := dgp.dataSourcesMap[verifRec.DataSourceID]
		curation := ds.Curation

		var dsID, matchCard, currCard, edDist, edDistStem int
		if m.MatchType != vlib.NoMatch {
			matchedCardinality := int(prsd.Cardinality)
			currentCardinality := int(prsdCurrent.Cardinality)

			dsID = verifRec.DataSourceID
			matchCard = matchedCardinality
			currCard = currentCardinality
			edDist = mItm.EditDistance
			edDistStem = mItm.EditDistanceStem
		}

		var outlink string
		if ds.OutlinkURL != "" && verifRec.OutlinkID.String != "" {
			outlink = strings.Replace(ds.OutlinkURL, "{}", verifRec.OutlinkID.String, 1)
		}

		var matchedCanonical, matchedCanonicalFull string
		if prsd.Parsed {
			matchedCanonical = prsd.Canonical.Simple
			matchedCanonicalFull = prsd.Canonical.Full
		}

		titleShort := ds.TitleShort
		if titleShort == "" {
			titleShort = ds.Title
		}

		resData := vlib.ResultData{
			RecordID:               verifRec.RecordID.String,
			LocalID:                verifRec.LocalID.String,
			Outlink:                outlink,
			DataSourceID:           dsID,
			DataSourceTitleShort:   titleShort,
			Curation:               curation,
			EntryDate:              ds.UpdatedAt,
			MatchedName:            verifRec.Name.String,
			MatchedCardinality:     matchCard,
			MatchedCanonicalSimple: matchedCanonical,
			MatchedCanonicalFull:   matchedCanonicalFull,
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
			ClassificationIDs:      verifRec.ClassificationIds.String,
			EditDistance:           edDist,
			StemEditDistance:       edDistStem,
			MatchType:              mItm.MatchType,
			ParsingQuality:         verifRec.ParseQuality,
		}

		mRec.MatchResults = append(mRec.MatchResults, &resData)

	}
	if discardedNum > 0 {
		log.Info().
			Msgf("Skipped %d low parsing quality names (e.g. '%s')",
				discardedNum,
				discardedExample,
			)
	}
	mRec.DataSourcesNum = len(sources)
}

func getVerifMap(vs []*dbshare.VerifSQL) map[string][]*dbshare.VerifSQL {
	vm := make(map[string][]*dbshare.VerifSQL)
	for _, v := range vs {
		if v.CanonicalID.String != "" {
			vm[v.CanonicalID.String] = append(vm[v.CanonicalID.String], v)
		} else {
			// for viruses
			vm[v.NameStringID.String] = append(vm[v.NameStringID.String], v)
		}
	}
	return vm
}

func (dgp *verifierpg) virusQuery(
	ctx context.Context,
	virMatches []*mlib.Match,
	input vlib.Input,
) ([]*dbshare.VerifSQL, error) {
	if len(virMatches) == 0 {
		return nil, nil
	}

	var res []*dbshare.VerifSQL
	ids := getUUIDs(virMatches)
	q := virusQ
	args := []interface{}{pq.Array(ids)}
	if len(input.DataSources) > 0 {
		args = append(args, pq.Array(input.DataSources))
		q += "\n    AND data_source_id = any($2::int[])"
	}

	err := sqlscan.Select(ctx, dgp.db, &res, q, args...)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (dgp verifierpg) nameQuery(
	ctx context.Context,
	canMatches []*mlib.Match,
	input vlib.Input,
) ([]*dbshare.VerifSQL, error) {

	var res []*dbshare.VerifSQL
	if len(canMatches) == 0 {
		return res, nil
	}

	ids := getUUIDs(canMatches)
	q := namesQ
	args := []interface{}{pq.Array(ids)}
	if len(input.DataSources) > 0 {
		args = append(args, pq.Array(input.DataSources))
		q += "\n    AND data_source_id = any($2::int[])"
	}

	err := sqlscan.Select(ctx, dgp.db, &res, q, args...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func getUUIDs(matches []*mlib.Match) []string {
	set := make(map[string]struct{})
	for _, v := range matches {
		for _, vv := range v.MatchItems {
			if vv.EditDistance > 5 {
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

// partitionMatches partitions matches into two categories:
// no match, match by canonical.
func partitionMatches(matches []mlib.Match) matchSplit {
	ms := matchSplit{
		noMatch:   make([]*mlib.Match, 0, len(matches)),
		virus:     make([]*mlib.Match, 0, len(matches)),
		canonical: make([]*mlib.Match, 0, len(matches)),
	}
	for i := range matches {
		switch matches[i].MatchType {
		case vlib.NoMatch:
			ms.noMatch = append(ms.noMatch, &matches[i])
		case vlib.Virus:
			ms.virus = append(ms.virus, &matches[i])
		default:
			ms.canonical = append(ms.canonical, &matches[i])
		}
	}
	return ms
}
