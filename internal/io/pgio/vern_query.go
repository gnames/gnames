package pgio

import (
	"context"
	"fmt"

	"github.com/gnames/gnames/pkg/ent/vern"
	"github.com/gnames/gnlib/ent/verifier"
	"github.com/jackc/pgx/v5"
)

func (p *pgio) GetVernaculars(ctx context.Context, records []vern.Record, langs []string) (map[vern.Record][]verifier.Vernacular, error) {

	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("pgio.GetVernaculars: create transaction failed: %w", err)
	}
	defer tx.Rollback(ctx)

	var res map[vern.Record][]verifier.Vernacular
	err = p.makeVernTemp(ctx, tx, records)
	if err != nil {
		return nil, fmt.Errorf("pgio.GetVernaculars: %w", err)
	}

	res, err = p.getVernaculars(ctx, tx, langs)
	if err != nil {
		return nil, fmt.Errorf("pgio.GetVernaculars: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("pgio.GetVernaculars: close transaction failed: %w", err)
	}
	return res, nil
}

func (p *pgio) makeVernTemp(ctx context.Context, tx pgx.Tx, recs []vern.Record) error {
	// NOTE: the temp table gets removed when the transaction is commited
	q := "CREATE TEMP TABLE temp_vern_records (data_source_id int, record_id varchar(256)) ON COMMIT DROP"
	_, err := tx.Exec(ctx, q)
	if err != nil {
		return fmt.Errorf("pgio.makeVernTemp: %w", err)
	}
	data := make([][]any, len(recs))
	for i, v := range recs {
		data[i] = []any{v.DataSourceID, v.RecordID}
	}

	rows := pgx.CopyFromRows(data)

	cols := []string{"data_source_id", "record_id"}
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"temp_vern_records"},
		cols,
		rows,
	)
	if err != nil {
		return fmt.Errorf("pgio.makeVernTemp: %w", err)
	}

	return nil
}

func (p *pgio) getVernaculars(
	ctx context.Context,
	tx pgx.Tx,
	langs []string,
) (map[vern.Record][]verifier.Vernacular, error) {
	res := make(map[vern.Record][]verifier.Vernacular)
	if len(langs) == 0 {
		return res, nil
	}
	q := `
SELECT vs.name, vsi.data_source_id, vsi.record_id, vsi.language, vsi.lang_code, vsi.country_code
	FROM vernacular_string_indices vsi
	  JOIN temp_vern_records vr ON vr.data_source_id = vsi.data_source_id AND vr.record_id = vsi.record_id
	  JOIN vernacular_strings vs ON vs.id = vsi.vernacular_string_id
	WHERE $1 = 'all' OR vsi.lang_code = ANY($2::text[])
	`
	allLangs := langs[0]
	rows, err := tx.Query(ctx, q, allLangs, langs)
	if err != nil {
		return res, fmt.Errorf("pgio.getVernaculars: finding vernaculars failed: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var vrn verifier.Vernacular
		var k vern.Record

		err = rows.Scan(
			&vrn.Name, &k.DataSourceID, &k.RecordID, &vrn.Language, &vrn.LanguageCode,
			&vrn.Country,
		)
		if err != nil {
			return res, fmt.Errorf("pgio.getVernaculars: scanning vernacular failed: %w", err)
		}
		res[k] = append(res[k], vrn)
	}
	return res, nil
}
