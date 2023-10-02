package verifierpg

import (
	"context"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/sqlscan"
	"github.com/gnames/gnames/internal/io/dbshare"
	"github.com/gnames/gnames/pkg/ent/verifier"
	vlib "github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnparser"
	"github.com/gnames/gnparser/ent/parsed"
	"github.com/lib/pq"
)

var idQ = fmt.Sprintf(`
SELECT %s
FROM verification v
WHERE name_string_id = $1
	`, dbshare.QueryFields)

func (vrf verifierpg) NameByID(
	inp vlib.NameStringInput,
) (*verifier.MatchRecord, error) {
	var res *verifier.MatchRecord
	cfg := gnparser.NewConfig(gnparser.OptWithDetails(true))
	parser := gnparser.New(cfg)

	verif, err := vrf.idQuery(inp)
	if err != nil {
		return nil, err
	}
	res = vrf.idData(parser, verif)
	return res, nil
}

func (vrf verifierpg) NameStringByID(id string) (string, error) {
	var res string
	row :=
		vrf.db.QueryRow("SELECT name FROM name_strings WHERE id = $1", id)
	err := row.Scan(&res)
	return res, err
}

func (vrf verifierpg) idData(
	parser gnparser.GNparser,
	match []*dbshare.VerifSQL,
) *verifier.MatchRecord {
	if len(match) == 0 || match[0].Name.String == "" {
		return nil
	}

	res := &verifier.MatchRecord{
		ID:       match[0].NameStringID.String,
		Name:     match[0].Name.String,
		Overload: len(match) > 20,
	}

	prsd := parser.ParseName(match[0].Name.String)
	if !prsd.Parsed && prsd.Virus {
		for _, v := range match {
			resData := vrf.addVirusMatch(v)
			resData.MatchType = vlib.Virus
			res.MatchResults = append(res.MatchResults, &resData)
		}
	}

	if !prsd.Parsed {
		return res
	}

	authors, year := dbshare.ProcessAuthorship(prsd.Authorship)
	res.Authors = authors
	res.Year = year
	res.CanonicalFull = prsd.Canonical.Full
	res.CanonicalSimple = prsd.Canonical.Simple
	res.Cardinality = prsd.Cardinality

	for _, v := range match {
		resData := vrf.addMatch(v, parser, prsd)
		resData.MatchType = vlib.Exact
		res.MatchResults = append(res.MatchResults, &resData)
	}
	return res
}

func (vrf verifierpg) addVirusMatch(
	vsql *dbshare.VerifSQL,
) vlib.ResultData {
	ds, outlink, title := vrf.dsData(vsql)
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
		ClassificationPath:   vsql.Classification.String,
		ClassificationRanks:  vsql.ClassificationRanks.String,
		ClassificationIDs:    vsql.ClassificationIds.String,
	}
	return resData
}

func (vrf verifierpg) dsData(
	vsql *dbshare.VerifSQL,
) (*vlib.DataSource, string, string) {
	var outlink string
	ds := vrf.dataSourcesMap[vsql.DataSourceID]
	if ds.OutlinkURL != "" && vsql.OutlinkID.String != "" {
		outlink = strings.Replace(ds.OutlinkURL, "{}", vsql.OutlinkID.String, 1)
	}
	titleShort := ds.TitleShort
	if titleShort == "" {
		titleShort = ds.Title
	}

	return ds, outlink, titleShort
}

func (vrf verifierpg) addMatch(
	vsql *dbshare.VerifSQL,
	p gnparser.GNparser,
	prsd parsed.Parsed,
) vlib.ResultData {
	var currName, currID, currRecordID string
	var currentCan, currentCanFull, outlink string
	var prsdCurrent parsed.Parsed
	if vsql.AcceptedRecordID.Valid {
		currRecordID = vsql.AcceptedRecordID.String
		currID = vsql.AcceptedNameID.String
		currName = vsql.AcceptedName.String
		prsdCurrent = p.ParseName(currName)
		if prsdCurrent.Parsed {
			currentCan = prsdCurrent.Canonical.Simple
			currentCanFull = prsdCurrent.Canonical.Full
		}
	}
	authors, year := dbshare.ProcessAuthorship(prsd.Authorship)
	ds, outlink, title := vrf.dsData(vsql)
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
		IsSynonym:              vsql.RecordID != vsql.AcceptedRecordID,
		ClassificationPath:     vsql.Classification.String,
		ClassificationRanks:    vsql.ClassificationRanks.String,
		ClassificationIDs:      vsql.ClassificationIds.String,
		EditDistance:           edDist,
		StemEditDistance:       edDistStem,
		ParsingQuality:         vsql.ParseQuality,
	}
	return resData
}

func (vrf verifierpg) idQuery(
	inp vlib.NameStringInput,
) ([]*dbshare.VerifSQL, error) {
	var res []*dbshare.VerifSQL
	q := idQ
	args := []interface{}{inp.ID}

	if len(inp.DataSources) > 0 {
		args = append(args, pq.Array(inp.DataSources))
		q += "\n    AND data_source_id = any($2::int[])"
	}
	err := sqlscan.Select(context.Background(), vrf.db, &res, q, args...)
	return res, err
}
