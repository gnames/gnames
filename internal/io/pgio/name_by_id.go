package pgio

import (
	"context"
	"fmt"

	"github.com/gnames/gnames/pkg/ent/verif"
	vlib "github.com/gnames/gnlib/ent/verifier"
)

func idQuery(inp vlib.NameStringInput) (string, []any) {
	q := fmt.Sprintf(`
SELECT %s
FROM verification v
WHERE name_string_id = $1
	`, queryFields)

	args := []any{inp.ID}

	if len(inp.DataSources) > 0 {
		args = append(args, inp.DataSources)
		q += "\n    AND data_source_id = any($2::int[])"
	}
	return q, args
}

func (p *pgio) idQueryRun(
	ctx context.Context,
	q string,
	args []any,
) ([]*verifSQL, error) {
	var res []*verifSQL
	rows, err := p.db.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("pgio.idQueryRun: %w", err)
	}
	defer rows.Close()

	res, err = rowsToVerifSQL(rows)
	if err != nil {
		return nil, fmt.Errorf("pgio.idQueryRun: %w", err)
	}
	return res, nil
}

func (p *pgio) idData(match []*verifSQL) *verif.MatchRecord {
	if len(match) == 0 || match[0].Name.String == "" {
		return nil
	}

	gnp := <-p.gnpPool
	defer func() { p.gnpPool <- gnp }()

	res := &verif.MatchRecord{
		ID:       match[0].NameStringID.String,
		Name:     match[0].Name.String,
		Overload: len(match) > 20,
	}

	prsd := gnp.ParseName(match[0].Name.String)
	if prsd.Virus {
		for _, v := range match {
			resData := p.addVirusMatch(v)
			resData.MatchType = vlib.Virus
			res.MatchResults = append(res.MatchResults, &resData)
		}
	}

	if !prsd.Parsed {
		return res
	}

	authors, year := processAuthorship(prsd.Authorship)
	res.Authors = authors
	res.Year = year
	res.CanonicalFull = prsd.Canonical.Full
	res.CanonicalSimple = prsd.Canonical.Simple
	res.Cardinality = prsd.Cardinality

	for _, v := range match {
		resData := p.addMatch(v, gnp, prsd)
		resData.MatchType = vlib.Exact
		res.MatchResults = append(res.MatchResults, &resData)
	}
	return res
}
