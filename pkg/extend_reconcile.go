package gnames

import (
	"strings"

	"github.com/gnames/gnames/pkg/ent/recon"
	"github.com/gnames/gnfmt"
	"github.com/gnames/gnlib/ent/reconciler"
	vlib "github.com/gnames/gnlib/ent/verifier"
)

func (g gnames) ExtendReconcile(q reconciler.ExtendQuery) (reconciler.ExtendOutput, error) {
	enc := gnfmt.GNjson{Pretty: false}
	rows := make(map[string]map[string][]reconciler.PropertyValue)
	var props []reconciler.Property
	for _, v := range q.Properties {
		prop := recon.NewProp(v.ID)
		if prop == recon.Unknown {
			continue
		}
		props = append(props, prop.Property())
	}
	res := reconciler.ExtendOutput{
		Meta: props,
		Rows: rows,
	}
	propRes := make(map[string]string)
	for _, v := range q.IDs {
		ns, err := g.NameByID(vlib.NameStringInput{
			ID:             v,
			WithAllMatches: true,
		}, true)
		if err != nil {
			return res, err
		}
		// should not happen during reconciliation
		if len(ns.Results) == 0 {
			continue
		}

		dataSourcesDet := getDataSourcesDetails(ns.Results)
		var jsn []byte
		jsn, err = enc.Encode(dataSourcesDet)
		var dataSourcesDetJSON string
		if err == nil {
			dataSourcesDetJSON = string(jsn)
		}

		bestResult := ns.Results[0]
		propRes[recon.CanonicalForm.Property().ID] =
			bestResult.MatchedCanonicalSimple
		propRes[recon.CurrentName.Property().ID] = bestResult.CurrentName
		propRes[recon.Classification.Property().ID] = jsonClassification(
			bestResult.ClassificationPath, bestResult.ClassificationRanks)
		propRes[recon.DataSource.Property().ID] =
			bestResult.DataSourceTitleShort
		propRes[recon.OutlinkURL.Property().ID] = bestResult.Outlink
		propRes[recon.AllDataSources.Property().ID] = dataSourcesDetJSON
		row := extensionRow(q.Properties, propRes)
		res.Rows[v] = row
		clear(propRes)
	}
	return res, nil
}

type hierarchy struct {
	Taxon string `json:"taxon"`
	Rank  string `json:"rank"`
}

func jsonClassification(cl string, rnk string) string {
	enc := gnfmt.GNjson{Pretty: false}
	taxa := strings.Split(cl, "|")
	ranks := strings.Split(rnk, "|")
	hr := make([]hierarchy, len(taxa))
	for i := range taxa {
		hr[i] = hierarchy{Taxon: taxa[i], Rank: ranks[i]}
	}
	res, _ := enc.Encode(hr)
	return string(res)
}

func getDataSourcesDetails(
	rs []*vlib.ResultData,
) map[string]vlib.DataSourceDetails {
	resMap := make(map[int]vlib.DataSourceDetails)
	for _, v := range rs {
		if strings.HasPrefix(v.RecordID, "gn_") ||
			// bug in Arctos
			strings.HasPrefix(v.RecordID, "'gn") {
			v.RecordID = ""
		}
		match := vlib.MatchShort{
			RecordID:   v.RecordID,
			NameString: v.MatchedName,
			AuthScore:  v.ScoreDetails.AuthorMatchScore > 0,
			Outlink:    v.Outlink,
		}
		if !match.AuthScore {
			continue
		}

		if _, ok := resMap[v.DataSourceID]; ok {
			continue
		} else {
			resMap[v.DataSourceID] = vlib.DataSourceDetails{
				DataSourceID: v.DataSourceID,
				TitleShort:   v.DataSourceTitleShort,
				Match:        match,
			}
		}
	}
	res := make(map[string]vlib.DataSourceDetails)
	for _, v := range resMap {
		res[v.TitleShort] = v
	}
	return res
}

func extensionRow(
	props []reconciler.Property,
	propsRes map[string]string,
) map[string][]reconciler.PropertyValue {
	res := make(map[string][]reconciler.PropertyValue)
	for _, v := range props {
		if _, ok := propsRes[v.ID]; ok {
			res[v.ID] = []reconciler.PropertyValue{{Str: propsRes[v.ID]}}
		}
	}
	return res
}
