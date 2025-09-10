package pgio

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/gnames/gnames/pkg/ent/verif"
	mlib "github.com/gnames/gnlib/ent/matcher"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnparser"
	"github.com/gnames/gnparser/ent/parsed"
	"github.com/pkg/errors"
)

const (
	// resultsThreshold is the number of returned results for a match after
	// which we remove results with worst ParsingQuality. This step allows
	// to get rid of names of bacterial strains, 'sec.' names etc.
	resultsThreshold = 100
)

var queryFields = `
  v.canonical_id, v.name, v.data_source_id, v.record_id,
  v.name_string_id, v.local_id, v.outlink_id, v.accepted_record_id,
  v.accepted_name_id, v.accepted_name, v.classification,
  v.classification_ranks, v.classification_ids, v.parse_quality
`

var namesQ = fmt.Sprintf(`
SELECT %s
FROM verification v
WHERE canonical_id = ANY($1::uuid[])
`, queryFields)

var virusQ = fmt.Sprintf(`
SELECT %s
FROM verification v
  WHERE name_string_id = ANY($1::uuid[])
`, queryFields)

func (p *pgio) nameQuery(
	ctx context.Context,
	canMatches []*mlib.Match,
	input vlib.Input,
) ([]*verifSQL, error) {

	var res []*verifSQL
	if len(canMatches) == 0 {
		return res, nil
	}

	ids := getUUIDs(canMatches)
	q := namesQ
	args := []any{ids}
	if len(input.DataSources) > 0 {
		args = append(args, input.DataSources)
		q += "\n    AND data_source_id = any($2::int[])"
	}

	rows, err := p.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("pgio.nameQuery: %w", err)
	}
	defer rows.Close()
	res, err = rowsToVerifSQL(rows)
	if err != nil {
		return nil, fmt.Errorf("pgio.nameQuery: %w", err)
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

func (p *pgio) virusQuery(
	ctx context.Context,
	virMatches []*mlib.Match,
	input vlib.Input,
) ([]*verifSQL, error) {
	var res []*verifSQL

	if len(virMatches) == 0 {
		return nil, nil
	}

	ids := getUUIDs(virMatches)
	q := virusQ
	args := []any{ids}
	if len(input.DataSources) > 0 {
		args = append(args, input.DataSources)
		q += "\n    AND data_source_id = any($2::int[])"
	}

	rows, err := p.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("pgio.virusQuery: %w", err)
	}
	defer rows.Close()

	res, err = rowsToVerifSQL(rows)
	if err != nil {
		return nil, fmt.Errorf("pgio.virusQuery: %w", err)
	}

	return res, nil
}

func (p *pgio) produceResultData(
	ms verif.MatchSplit,
	parser gnparser.GNparser,
	verCan []*verifSQL,
	verVir []*verifSQL,
) map[string]*verif.MatchRecord {

	// deal with NoMatch first
	allMatchRecs := make(map[string]*verif.MatchRecord)
	for _, v := range ms.NoMatch {
		allMatchRecs[v.ID] = &verif.MatchRecord{
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
	for _, match := range ms.Virus {
		mr := verif.MatchRecord{
			ID:       match.ID,
			Name:     match.Name,
			Overload: len(match.MatchItems) > 20,
		}
		for _, mi := range match.MatchItems {
			p.populateVirusMatchRecord(mi, *match, &mr, verifMap)
		}
		allMatchRecs[match.ID] = &mr
	}

	// deal with Canonicals
	for _, match := range ms.Canonical {
		// TODO check if parsing affects speed too much
		prsd := parser.ParseName(match.Name)

		if !prsd.Parsed {
			slog.Error("Cannot parse (shold never happen)",
				"error", errors.New("cannot parse"),
				slog.String("name", match.Name),
			)
		}
		authors, year := processAuthorship(prsd.Authorship)

		mr := verif.MatchRecord{
			ID:              match.ID,
			Name:            match.Name,
			Cardinality:     int(prsd.Cardinality),
			CanonicalSimple: prsd.Canonical.Simple,
			CanonicalFull:   prsd.Canonical.Full,
			Authors:         authors,
			Year:            year,
		}

		for _, mi := range match.MatchItems {
			p.populateMatchRecord(mi, *match, &mr, parser, verifMap)
		}
		allMatchRecs[match.ID] = &mr
	}

	return allMatchRecs
}

func (p *pgio) populateVirusMatchRecord(
	mi mlib.MatchItem,
	m mlib.Match,
	mr *verif.MatchRecord,
	verifMap map[string][]*verifSQL,
) error {
	verifRecs, ok := verifMap[mi.ID]
	if !ok {
		return fmt.Errorf("no match for %s", mi.ID)
	}

	for _, vsql := range verifRecs {
		resData := p.addVirusMatch(vsql)
		resData.MatchType = m.MatchType

		mr.MatchResults = append(mr.MatchResults, &resData)
	}
	return nil
}

func (p *pgio) populateMatchRecord(
	mItm mlib.MatchItem,
	m mlib.Match,
	mRec *verif.MatchRecord,
	parser gnparser.GNparser,
	verifMap map[string][]*verifSQL,
) {
	verifRecs, ok := verifMap[mItm.ID]
	if !ok {
		return
	}

	recsNum := len(verifRecs)
	var discardedExample string
	var discardedNum int
	for _, vsql := range verifRecs {
		// if there is a lot of records, most likely many of them are surrogates
		// that parser is not able to catch. Surrogates would parse with worst
		// parsing quality (4)
		mRec.Overload = recsNum > resultsThreshold
		if recsNum > resultsThreshold && vsql.ParseQuality == 4 {
			if discardedExample == "" {
				discardedExample = vsql.Name.String
			}
			discardedNum++
			continue
		}

		prsd := parser.ParseName(vsql.Name.String)
		resData := p.addMatch(vsql, parser, prsd)

		resData.MatchType = mItm.MatchType
		resData.EditDistance = mItm.EditDistance
		resData.StemEditDistance = mItm.EditDistanceStem

		mRec.MatchResults = append(mRec.MatchResults, &resData)
	}
	if discardedNum > 0 {
		slog.Warn("Skipped low parsing quality names",
			slog.String("example", discardedExample),
			slog.Int("skippedNum", discardedNum),
		)
	}
}

func getVerifMap(vs []*verifSQL) map[string][]*verifSQL {
	vm := make(map[string][]*verifSQL)
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

func (p *pgio) dsData(
	vsql *verifSQL,
) (*vlib.DataSource, string, string) {
	var outlink string
	ds := p.dsm[vsql.DataSourceID]
	if ds.OutlinkURL != "" && vsql.OutlinkID.String != "" {
		outlink = strings.Replace(ds.OutlinkURL, "{}", vsql.OutlinkID.String, 1)
	}
	titleShort := ds.TitleShort
	if titleShort == "" {
		titleShort = ds.Title
	}

	return ds, outlink, titleShort
}

func (p *pgio) addVirusMatch(
	vsql *verifSQL,
) vlib.ResultData {
	ds, outlink, title := p.dsData(vsql)
	hasTaxonData := p.hasTaxonData(vsql)
	resData := vlib.ResultData{
		RecordID:             vsql.RecordID.String,
		LocalID:              vsql.LocalID.String,
		Outlink:              outlink,
		DataSourceID:         vsql.DataSourceID,
		DataSourceTitleShort: title,
		Curation:             ds.Curation,
		EntryDate:            ds.UpdatedAt,
		MatchedNameID:        vsql.NameStringID.String,
		MatchedName:          vsql.Name.String,
		TaxonomicStatus:      getTaxonomicStatus(vsql, hasTaxonData),
		ClassificationPath:   vsql.Classification.String,
		ClassificationRanks:  vsql.ClassificationRanks.String,
		ClassificationIDs:    vsql.ClassificationIds.String,
	}
	resData.IsSynonym = resData.TaxonomicStatus == vlib.SynonymTaxStatus
	return resData
}

func (p *pgio) addMatch(
	vsql *verifSQL,
	gnp gnparser.GNparser,
	prsd parsed.Parsed,
) vlib.ResultData {
	var currName, currID, currRecordID string
	var currentCan, currentCanFull, outlink string
	var prsdCurrent parsed.Parsed
	hasTaxonData := p.hasTaxonData(vsql)
	status := getTaxonomicStatus(vsql, hasTaxonData)
	if hasTaxonData {
		currRecordID = vsql.AcceptedRecordID.String
		currID = vsql.AcceptedNameID.String
		currName = vsql.AcceptedName.String
		prsdCurrent = gnp.ParseName(currName)
		if prsdCurrent.Parsed {
			currentCan = prsdCurrent.Canonical.Simple
			currentCanFull = prsdCurrent.Canonical.Full
		}
	}
	authors, year := processAuthorship(prsd.Authorship)
	ds, outlink, title := p.dsData(vsql)
	var dsID, matchCard, currCard, edDist, edDistStem int

	matchedCardinality := int(prsd.Cardinality)
	currentCardinality := int(prsdCurrent.Cardinality)

	dsID = vsql.DataSourceID
	matchCard = matchedCardinality
	currCard = currentCardinality

	var matchedCanonical, matchedCanonicalFull string
	matchedCanonical = prsd.Canonical.Simple
	matchedCanonicalFull = prsd.Canonical.Full

	resData := vlib.ResultData{
		RecordID:               vsql.RecordID.String,
		LocalID:                vsql.LocalID.String,
		Outlink:                outlink,
		DataSourceID:           dsID,
		DataSourceTitleShort:   title,
		Curation:               ds.Curation,
		EntryDate:              ds.UpdatedAt,
		MatchedNameID:          vsql.NameStringID.String,
		MatchedName:            vsql.Name.String,
		MatchedCardinality:     matchCard,
		MatchedCanonicalSimple: matchedCanonical,
		MatchedCanonicalFull:   matchedCanonicalFull,
		MatchedAuthors:         authors,
		MatchedYear:            year,
		CurrentRecordID:        currRecordID,
		CurrentNameID:          currID,
		CurrentName:            currName,
		CurrentCardinality:     currCard,
		CurrentCanonicalSimple: currentCan,
		CurrentCanonicalFull:   currentCanFull,
		TaxonomicStatus:        status,
		ClassificationPath:     vsql.Classification.String,
		ClassificationRanks:    vsql.ClassificationRanks.String,
		ClassificationIDs:      vsql.ClassificationIds.String,
		EditDistance:           edDist,
		StemEditDistance:       edDistStem,
		ParsingQuality:         vsql.ParseQuality,
	}
	resData.IsSynonym = resData.TaxonomicStatus == vlib.SynonymTaxStatus
	return resData
}

func (p pgio) hasTaxonData(vsql *verifSQL) bool {
	var res bool
	if _, ok := p.dsm[vsql.DataSourceID]; ok {
		res = p.dsm[vsql.DataSourceID].HasTaxonData
	}
	return res
}

func getTaxonomicStatus(vsql *verifSQL, hasTaxonData bool) vlib.TaxonomicStatus {
	if strings.TrimSpace(vsql.Classification.String) == "" {
		return vlib.UnknownTaxStatus
	}
	if vsql.RecordID != vsql.AcceptedRecordID {
		return vlib.SynonymTaxStatus
	}
	if hasTaxonData {
		return vlib.AcceptedTaxStatus
	}
	return vlib.UnknownTaxStatus
}
