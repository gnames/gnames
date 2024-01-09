package pgio

import (
	"fmt"
	"strings"

	"github.com/gnames/gnparser/ent/parsed"
	"github.com/gnames/gnquery/ent/search"
)

func spQuery(
	inp search.Input,
	spWordIDs []int,
	spWord string,
) (string, []any) {
	// prepare species epithet word for SQL
	sp, insertStr := prepareSpWord(spWord)
	args := []any{sp, spWordIDs}
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
	// prepare genus word for SQL
	gen := prepareGenWord(inp)
	if gen != "" {
		args = append(args, gen)
		genQ := fmt.Sprintf("      AND c.name LIKE $%d", len(args))
		spQ = append(spQ, genQ)
	}

	// add higher taxon constraint
	if tx := inp.ParentTaxon; tx != "" {
		tx = "%" + tx + "%"
		args = append(args, tx)
		clQ := fmt.Sprintf("      AND v.classification LIKE $%d", len(args))
		spQ = append(spQ, clQ)
	}

	spQ = append(spQ, ")")
	res := strings.Join(spQ, "\n")
	return res, args
}

// prepareSpWord prepares a species epithet word to be compatible with
// SQL.
func prepareSpWord(spWord string) (string, string) {
	iStr := "normalized like $1"
	bs := []byte(spWord)
	l := len(bs)
	if bs[l-1] == '.' {
		bs[l-1] = '%'
		return string(bs), iStr
	}

	iStr = "modified = $1"
	st := parsed.NormalizeByType(spWord, parsed.SpEpithetType)
	return st, iStr
}

// prepareGenWord prepares a genus epithet word to be compatible with
// SQL.
func prepareGenWord(inp search.Input) string {
	g := inp.Genus
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
