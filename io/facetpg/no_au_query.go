package facetpg

import (
	"fmt"

	"github.com/gnames/gnames/io/internal/dbshare"
)

func (f *facetpg) noAuQuery(
	q string,
	args []interface{},
) (string, []interface{}) {
	noAuQ := fmt.Sprintf(`
SELECT distinct %s FROM verification v
  RIGHT JOIN sp ON v.name_string_id = sp.name_string_id
  WHERE 1=1`, dbshare.QueryFields)

	noAuQ, args = f.queryEnd(noAuQ, args)
	noAuQ = q + noAuQ

	return noAuQ, args
}
