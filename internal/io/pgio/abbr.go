package pgio

import (
	"context"
	"fmt"
	"sync"
	"unicode/utf8"

	"github.com/gnames/gnames/pkg/ent/abbr"
)

// prefixSizes caches sizes of genera by prefix. They change only when
// the database is rebuilt.
var prefixSizes sync.Map

// AbbrLookups returns rows of `abbr_lookups` for the given keys.
func (p *pgio) AbbrLookups(
	ctx context.Context,
	keys []string,
) ([]abbr.Lookup, error) {
	q := `
SELECT abbr_key, genus, canonical_stem_id::text, canonical_stem,
       coalesce(ds_num, 0)
FROM abbr_lookups
WHERE abbr_key = ANY($1)`

	rows, err := p.db.Query(ctx, q, keys)
	if err != nil {
		return nil, fmt.Errorf("pgio.AbbrLookups: %w", err)
	}
	defer rows.Close()

	var res []abbr.Lookup
	for rows.Next() {
		var l abbr.Lookup
		var dsNum int16
		err = rows.Scan(
			&l.Key, &l.Genus, &l.CanonicalStemID, &l.CanonicalStem, &dsNum,
		)
		if err != nil {
			return nil, fmt.Errorf("pgio.AbbrLookups: %w", err)
		}
		l.DsNum = int(dsNum)
		res = append(res, l)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("pgio.AbbrLookups: %w", err)
	}
	return res, nil
}

// AbbrGenera returns rows of `abbr_genera` for the given genera.
func (p *pgio) AbbrGenera(
	ctx context.Context,
	genera []string,
) ([]abbr.GenusRecord, error) {
	q := `
SELECT genus, kingdom, epithets_num, names_num,
       coalesce(classification, ''), coalesce(classification_ranks, ''),
       coalesce(data_source_id, 0), known
FROM abbr_genera
WHERE genus = ANY($1)`

	rows, err := p.db.Query(ctx, q, genera)
	if err != nil {
		return nil, fmt.Errorf("pgio.AbbrGenera: %w", err)
	}
	defer rows.Close()

	var res []abbr.GenusRecord
	for rows.Next() {
		var r abbr.GenusRecord
		var epithetsNum, namesNum, dsID int64
		err = rows.Scan(
			&r.Genus, &r.Kingdom, &epithetsNum, &namesNum,
			&r.Classification, &r.ClassificationRanks, &dsID, &r.Known,
		)
		if err != nil {
			return nil, fmt.Errorf("pgio.AbbrGenera: %w", err)
		}
		r.EpithetsNum = int(epithetsNum)
		r.NamesNum = int(namesNum)
		r.DataSourceID = int(dsID)
		res = append(res, r)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("pgio.AbbrGenera: %w", err)
	}
	return res, nil
}

// AbbrPrefixSize returns the sum of sizes of genera that start with the
// prefix. Homonyms have the same size in every row, so each genus counts
// once.
func (p *pgio) AbbrPrefixSize(
	ctx context.Context,
	prefix string,
) (int, error) {
	if v, ok := prefixSizes.Load(prefix); ok {
		return v.(int), nil
	}

	q := `
SELECT coalesce(sum(epithets_num), 0)
FROM (
  SELECT DISTINCT genus, epithets_num
  FROM abbr_genera
  WHERE left(genus, $2) = $1
) g`

	var res int64
	err := p.db.QueryRow(
		ctx, q, prefix, utf8.RuneCountInString(prefix),
	).Scan(&res)
	if err != nil {
		return 0, fmt.Errorf("pgio.AbbrPrefixSize: %w", err)
	}
	prefixSizes.Store(prefix, int(res))
	return int(res), nil
}

// AbbrCanonicals returns canonical forms for IDs of stemmed canonical
// forms. Only names with epithets are returned, as in `abbr_lookups`.
func (p *pgio) AbbrCanonicals(
	ctx context.Context,
	stemIDs []string,
) (map[string][]string, error) {
	q := `
SELECT DISTINCT ns.canonical_stem_id::text, c.name
FROM name_strings ns
JOIN canonicals c ON c.id = ns.canonical_id
WHERE ns.canonical_stem_id = ANY($1::uuid[])
  AND ns.cardinality > 1`

	rows, err := p.db.Query(ctx, q, stemIDs)
	if err != nil {
		return nil, fmt.Errorf("pgio.AbbrCanonicals: %w", err)
	}
	defer rows.Close()

	res := make(map[string][]string)
	for rows.Next() {
		var id, name string
		if err = rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("pgio.AbbrCanonicals: %w", err)
		}
		res[id] = append(res[id], name)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("pgio.AbbrCanonicals: %w", err)
	}
	return res, nil
}
