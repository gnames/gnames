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
		spQ = append(spQ, "      AND c.name LIKE $3")
		args = append(args, gen)
	}

	tx, ds := f.prepareTxWord()
	if ds > 0 {
		args = append(args, ds)
		spQ = append(spQ, "      AND v.data_source_id = $4")
		if tx != "" {
			args = append(args, tx)
			spQ = append(spQ, "      AND v.classification LIKE $5")
		}
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
	if len(g) < 3 {
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

func (f *facetpg) prepareTxWord() (string, int) {
	ds := f.DataSourceID
	tx := f.ParentTaxon

	if tx != "" {
		tx = "%" + tx + "%"
		if ds == 0 {
			ds = 1
		}
	}

	return tx, ds
}
