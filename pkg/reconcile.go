package gnames

import (
	"strconv"
	"strings"

	"github.com/gnames/gnames/pkg/ent/lexgroup"
	"github.com/gnames/gnames/pkg/ent/recon"
	"github.com/gnames/gnlib/ent/reconciler"
	vlib "github.com/gnames/gnlib/ent/verifier"
)

func (g gnames) Reconcile(
	verif vlib.Output,
	qs map[string]reconciler.Query,
	ids []string,
) reconciler.Output {
	res := reconciler.Output(make(map[string]reconciler.ReconciliationResult))

	for i, v := range verif.Names {
		prs := qs[ids[i]].Properties
		lgs := lexgroup.NameToLexicalGroups(v)
		lgs = filterLexGrpByProperties(lgs, prs)
		var rcs []reconciler.ReconciliationCandidate

		// the weights (cur, first, auth, etc) allow to group results into
		// clusters of scores, it makes reconciliation with OpenRefine easier
		// and faster.
		for i, lg := range lgs {
			cur := 0.95
			first := 0.9
			auth := 0.9
			exact := 0.70
			complete := 0.4
			oneCode := 0.80

			if i == 0 {
				first = 1
			}

			if lg.Data[0].MatchType == vlib.Exact {
				exact = 1
				complete = 1
			} else if lg.Data[0].MatchType == vlib.Fuzzy {
				complete = 1
			} else if lg.Data[0].MatchType == vlib.PartialExact {
				exact = 1
				if lg.Data[0].MatchedCardinality == 1 {
					complete = 0.2
				}
			} else if lg.Data[0].MatchType == vlib.PartialFuzzy {
				if lg.Data[0].MatchedCardinality == 1 {
					complete = 0.2
				}
			}

			if lg.Data[0].ScoreDetails.AuthorMatchScore > 0.2 {
				auth = 1
			}
			if lg.Data[0].ScoreDetails.CuratedDataScore > 0 {
				cur = 1
			}
			if len(lg.NomCodes) < 2 {
				oneCode = 1
			}

			score := first * auth * cur * oneCode * exact * complete

			// features supply data for the score calculation
			features := []reconciler.Feature{
				{ID: "first_result", Value: first},
				{ID: "exact_match", Value: exact},
				{ID: "complete_canonical_match", Value: complete},
				{ID: "authors_compatible", Value: auth},
				{ID: "has_curation_process", Value: cur},
				{ID: "single_nomenclatural_code", Value: oneCode},
			}

			rc := reconciler.ReconciliationCandidate{
				ID:       lg.ID,
				Score:    score,
				Match:    score == 1,
				Features: features,
				Name:     lg.Name,
			}
			rcs = append(rcs, rc)
		}
		res[ids[i]] = reconciler.ReconciliationResult{
			Result: rcs,
		}
	}
	return res
}

func filterLexGrpByProperties(
	lgs []lexgroup.LexicalGroup,
	prs []reconciler.PropertyInfo,
) []lexgroup.LexicalGroup {
	var res []lexgroup.LexicalGroup
	if len(prs) == 0 {
		return lgs
	}
	for i := range lgs {
		grp := filterGroup(lgs[i], prs)
		if len(grp.Data) > 0 {
			res = append(res, grp)
		}
	}
	return res
}

func filterGroup(
	lg lexgroup.LexicalGroup,
	prs []reconciler.PropertyInfo,
) lexgroup.LexicalGroup {
	var res lexgroup.LexicalGroup
	fs := make(map[string]string)
	for i := range prs {
		pid := strings.ToLower(prs[i].PropertyID)
		fs[pid] = prs[i].PropertyValue
	}
	if taxon, ok := fs[recon.HigherTaxon.Property().ID]; ok {
		res = filterByTaxon(lg, taxon)
	} else {
		res = lg
	}
	if idStr, ok := fs[recon.DataSourceIDs.Property().ID]; ok {
		ids := filteredDataSrcIDs(idStr)
		if len(ids) > 0 {
			res = filterByDataSource(lg, ids)
		}
	}
	return res
}

func filteredDataSrcIDs(s string) map[int]struct{} {
	res := make(map[int]struct{})
	elements := strings.Split(s, ",")
	for _, v := range elements {
		v = strings.TrimSpace(v)
		id, err := strconv.Atoi(v)
		if err == nil {
			res[id] = struct{}{}
		}
	}
	return res
}

func filterByTaxon(
	lg lexgroup.LexicalGroup,
	taxon string,
) lexgroup.LexicalGroup {
	var res lexgroup.LexicalGroup
	var ds []*vlib.ResultData
	d := lg.Data
	for i := range d {
		if d[i].ClassificationPath == "" {
			continue
		}
		if strings.Index(d[i].ClassificationPath, taxon) > -1 {
			ds = append(ds, d[i])
		}
	}
	if len(ds) > 0 {
		res = lexgroup.New(ds[0])
		res.Data = ds
	}
	return res
}

func filterByDataSource(
	lg lexgroup.LexicalGroup,
	ids map[int]struct{},
) lexgroup.LexicalGroup {
	var res lexgroup.LexicalGroup
	var ds []*vlib.ResultData

	d := lg.Data
	for i := range d {
		if _, ok := ids[d[i].DataSourceID]; ok {
			ds = append(ds, d[i])
		}
	}

	if len(ds) > 0 {
		res = lexgroup.New(ds[0])
		res.Data = ds
	}
	return res
}
