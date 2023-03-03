package facetpg

import (
	"fmt"

	"github.com/gnames/gnames/internal/io/dbshare"
	"github.com/gnames/gnparser/ent/parsed"
)

func (f *facetpg) auQuery(
	q string,
	args []interface{},
) (string, []interface{}) {
	var auStr string
	auStr, args = f.prepareAuWord(args)
	args = append(args, int(parsed.AuthorWordType))
	auQ := fmt.Sprintf(`
au AS (
  SELECT wc.name_string_id
    FROM word_name_strings wc
      JOIN words w ON w.id = wc.word_id
      JOIN sp ON sp.name_string_id = wc.name_string_id
    WHERE w.modified %s
    AND w.type_id = $%d
)
SELECT distinct %s
  FROM verification v
    RIGHT JOIN au ON v.name_string_id = au.name_string_id
    WHERE 1=1`, auStr, len(args), dbshare.QueryFields)

	auQ, args = f.queryEnd(auQ, args)
	auQ = q + "," + auQ

	return auQ, args
}

func (f *facetpg) prepareAuWord(
	args []interface{},
) (string, []interface{}) {
	var str string
	au := parsed.NormalizeByType(f.Author, parsed.AuthorWordType)
	bs := []byte(au)
	l := len(bs)
	if bs[l-1] == '.' {
		bs[l-1] = '%'
		args = append(args, string(bs))
		str = fmt.Sprintf("like $%d", len(args))
	} else {
		args = append(args, au)
		str = fmt.Sprintf("= $%d", len(args))
	}
	return str, args
}
