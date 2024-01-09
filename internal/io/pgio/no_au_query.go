package pgio

import (
	"fmt"

	"github.com/gnames/gnquery/ent/search"
)

func noAuQuery(
	q string,
	inp search.Input,
	args []interface{},
) (string, []interface{}) {
	noAuQ := fmt.Sprintf(`
SELECT distinct %s FROM verification v
  RIGHT JOIN sp ON v.name_string_id = sp.name_string_id
  WHERE 1=1`, queryFields)

	noAuQ, args = queryEnd(noAuQ, inp, args)
	noAuQ = q + noAuQ

	return noAuQ, args
}
