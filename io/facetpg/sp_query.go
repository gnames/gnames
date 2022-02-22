package facetpg

import (
	"fmt"
	"strings"

	"github.com/gnames/gnparser/ent/parsed"
	"github.com/lib/pq"
)

func (f *facetpg) spQuery() (string, []interface{}) {
	sp, insertStr := f.prepareSpWord()
	args := []interface{}{sp, pq.Array(f.spWordIDs)}
	q := fmt.Sprintf(`
WITH sp AS (
  SELECT DISTINCT v.name_string_id
    FROM word_name_strings wc
      JOIN words w ON w.id = wc.word_id
      JOIN canonicals c ON c.id = wc.canonical_id
      JOIN verification v ON c.id = v.canonical_id
    WHERE %s
      AND type_id = any($2::int[])`, insertStr,
	)
	spQ := []string{q}
	gen := f.prepareGenWord()
	if gen != "" {
		args = append(args, gen)
		genQ := fmt.Sprintf("      AND c.name LIKE $%d", len(args))
		spQ = append(spQ, genQ)
	}

	// add higher taxon constraint
	if tx := f.ParentTaxon; tx != "" {
    tx = "%" + tx + "%"
		args = append(args, tx)
		clQ := fmt.Sprintf("      AND v.classification LIKE $%d", len(args))
		spQ = append(spQ, clQ)
	}

	spQ = append(spQ, ")")
	res := strings.Join(spQ, "\n")
	return res, args
}

func (f *facetpg) prepareSpWord() (string, string) {
	iStr := "normalized like $1"
	bs := []byte(f.spWord)
	l := len(bs)
	if bs[l-1] == '.' {
		bs[l-1] = '%'
		return string(bs), iStr
	}

	iStr = "modified = $1"
	st := parsed.NormalizeByType(f.spWord, parsed.SpEpithetType)
	return st, iStr
}

func (f *facetpg) prepareGenWord() string {
	g := f.Genus
	if len(g) < 2 {
		return ""
	}

	bs := []byte(g)
	l := len(bs)
	if bs[l-1] == '.' {
		bs[l-1] = '%'
		return string(bs)
	}

	return g + " %"
}
