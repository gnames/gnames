// package lexgroup creates lexical groups out of matched
// results from verification.
package lexgroup

import (
	"cmp"
	"slices"
	"strings"

	"github.com/gnames/gnlib/ent/verifier"
	"github.com/gnames/gnparser"
	"github.com/gnames/gnparser/ent/parsed"
)

// LexicalGroup combines together name-strings which seem to belong to the
// same scientific name.
type LexicalGroup struct {
	// ID corresponds to ID of a matched name UUID v5.
	ID string

	// Name is the matched name.
	Name string

	// Score is the same as a sort score for a match.
	Score float64

	// NomCodes contains nomenclatural codes detected for this lexical group.
	NomCodes map[string]struct{}

	// AuthMatch is the score for author matching.
	AuthMatch float32

	// CanonicalFull is the full canonical form of the matched name.
	CanonicalFull string

	// LexicalVariants is a slice of all spellings compatible with this group.
	LexicalVariants []string

	// Data is a slice of all corresponding matched data.
	Data []*verifier.ResultData
}

// New creates new LexicalGroup instance out of a *verifier.ResultData.
func New(rd *verifier.ResultData) LexicalGroup {
	res := LexicalGroup{
		ID:              rd.MatchedNameID,
		Name:            rd.MatchedName,
		CanonicalFull:   rd.MatchedCanonicalFull,
		AuthMatch:       rd.ScoreDetails.AuthorMatchScore,
		Score:           rd.SortScore,
		NomCodes:        make(map[string]struct{}),
		LexicalVariants: []string{rd.MatchedName},
		Data:            []*verifier.ResultData{rd},
	}
	code := getCode(rd)
	if code != "" {
		res.NomCodes[code] = struct{}{}
	}
	return res
}

// record contains data required for matching by canonical form or by
// authorship
type record struct {
	// idx is the position of a result in a Name.Results,
	// the smaller is its valuei the better is the match.
	idx int
	p   parsed.Parsed
	au  *authors
	rd  *verifier.ResultData
}

// simplified data for authorship
type authors struct {
	isCombination bool
	orig          string
	comb          string
}

// group is used to create a lexical group. The `data` is the group,
// `au` is used to compare group authors.
type group struct {
	au   *authors
	data []record
}

// NameToLexicalGroups takes verification results for a name and reorganizes
// the matching results into lexical groups. Note that if only best result exit
// only one lexical groups with one mamber is returned.
func NameToLexicalGroups(n verifier.Name) []LexicalGroup {
	var res []LexicalGroup

	// return nil if no matches are found
	if n.MatchType == verifier.NoMatch {
		return res
	}

	// if there is only the best result, create a simple LexicalGroup from
	// just one item.
	if n.BestResult != nil {
		res = []LexicalGroup{New(n.BestResult)}
	}

	// if Results are empty, returns result, it can be LexicalGroup from
	// only BestResult, or nil
	if len(n.Results) == 0 {
		return res
	}

	// special case if a name is virus, return one lexical group
	if n.MatchType == verifier.Virus {
		return lexGroupVirus(n)
	}
	// in all other cases try to find all lexical variants
	return lexGroups(n)
}

// lexGroupVirus deals with a special case where the matches are virus
// name-strings.
func lexGroupVirus(n verifier.Name) []LexicalGroup {
	lg := New(n.Results[0])
	var names []string
	namesMap := make(map[string]struct{})
	rs := n.Results
	for i := range rs {
		if _, ok := namesMap[rs[i].MatchedName]; ok {
			continue
		}
		names = append(names, rs[i].MatchedName)
		namesMap[rs[i].MatchedName] = struct{}{}
	}
	lg.LexicalVariants = names
	return []LexicalGroup{lg}
}

// lexGroup converts verifier.Name to lexical groups.
func lexGroups(n verifier.Name) []LexicalGroup {
	p := gnparser.New(gnparser.NewConfig(gnparser.OptWithDetails(true)))

	// create records out of results
	ds := make([]record, len(n.Results))
	for i := range n.Results {
		parsed := p.ParseName(n.Results[i].MatchedName)
		ds[i] = record{
			idx: i,
			p:   parsed,
			au:  getAuthors(parsed),
			rd:  n.Results[i],
		}
	}

	var res []group
	// first see if simple canonical differ. It is quite possible
	// because matching happens between stemmed versions of names.
	gs := splitByCanonical(ds)

	if ds[0].rd.MatchedCardinality > 2 {
		// use ranks to distinguish between infraspecies
		gs = splitByFullCanonical(gs)
		res = append(res, gs...)

	} else {
		// use authorship to distinguish names with cardinality 1 or 2
		for i := range gs {
			gs2 := splitByAuthorship(gs[i])
			res = append(res, gs2...)
		}
	}

	// convert results into lexical groups
	return toLexicalGroups(res)
}

func getAuthors(p parsed.Parsed) *authors {
	var isComb bool
	var orig, comb []rune
	if !p.Parsed {
		return nil
	}

	if p.Authorship == nil || p.Authorship.Original == nil {
		return nil
	}

	if p.Authorship.Normalized[0] == '(' {
		isComb = true
	}

	for _, v := range p.Authorship.Original.Authors {
		words := strings.Split(v, " ")
		au := words[len(words)-1]
		orig = append(orig, []rune(au)[0])
	}

	if p.Authorship.Combination != nil {
		for _, v := range p.Authorship.Combination.Authors {
			words := strings.Split(v, " ")
			au := words[len(words)-1]
			comb = append(comb, []rune(au)[0])

		}
	}
	res := &authors{
		isCombination: isComb,
		orig:          string(orig),
		comb:          string(comb),
	}
	return res
}

func splitByFullCanonical(gs []group) []group {
	var res []group
	for i := range gs {
		mp := make(map[string][]record)
		for j := range gs[i].data {
			rd := gs[i].data[j].rd
			can := rd.MatchedCanonicalFull
			mp[can] = append(mp[can], gs[i].data[j])
		}

		for _, v := range mp {
			g := toGroup(v)
			gs := splitByAuthorship(g)
			res = append(res, gs...)
		}
	}
	return res
}

func toGroup(d []record) group {
	res := group{
		data: d,
	}
	return res
}

func splitByCanonical(ds []record) []group {
	cans := make(map[string][]record)
	for i := range ds {
		can := ds[i].rd.MatchedCanonicalSimple
		if _, ok := cans[can]; ok {
			cans[can] = append(cans[can], ds[i])
		} else {
			cans[can] = []record{ds[i]}
		}
	}
	var res []group
	for _, v := range cans {
		res = append(res, group{data: v})
	}
	return res
}

func splitByAuthorship(g group) []group {
	var res []group
	gs := matchByOrig(g)
	for _, v := range gs {
		gs := matchByCombo(v)
		res = append(res, gs...)
	}
	return res
}

func matchByOrig(g group) []group {
	var res []group
	gmap := make(map[string][]record)
	var noAu []record
	for _, v := range g.data {
		au := auToString(v.au, false)
		if au == "" {
			noAu = append(noAu, v)
			continue
		}
		if _, ok := gmap[au]; ok {
			gmap[au] = append(gmap[au], v)
		} else {
			gmap[au] = []record{v}
		}
	}
	for _, v := range gmap {
		res = append(res, group{data: v})
	}
	// if len(res) == 0 && len(noAu) > 0 {
	if len(noAu) > 0 {
		res = append(res, group{data: noAu})
	}
	return res
}

func matchByCombo(g group) []group {
	var res []group
	tmp := make(map[string][]record)
	var noComb []record
	for _, v := range g.data {
		var cmb string
		if v.au != nil {
			cmb = v.au.comb
		}
		if cmb == "" {
			noComb = append(noComb, v)
			continue
		}
		tmp[cmb] = append(tmp[cmb], v)
	}
	if len(tmp) == 0 {
		gr := group{data: noComb}
		return []group{gr}

	}
	for _, v := range tmp {
		for _, vv := range noComb {
			v = append(v, vv)
		}
		res = append(res, group{data: v})
	}
	return res
}

// simplify authors into a string for matching
func auToString(as *authors, withCombo bool) string {
	var res string
	if as == nil {
		return res
	}
	if as.isCombination {
		res += "comb|"
	}
	res += as.orig + "|"
	if withCombo {
		res += as.comb
	}
	return res
}

func authorsMatch(a1, a2 *authors) bool {
	if a1 == nil || a2 == nil {
		return true
	}
	if a1.isCombination != a2.isCombination {
		return false
	}
	if a1.orig == a2.orig {
		return true
	}
	return false
}

func addByAuthorship(gs []group, d []record) []group {
	for _, v := range gs {
		setAuthorship(v)
		au := auToString(v.au, false)
		var isCmb bool
		var cmb string
		if v.au != nil {
			isCmb = v.au.isCombination
			cmb = v.au.comb
		}
		for i := range d {
			if d[i].p.Authorship == nil {
				v.data = append(v.data, d[i])
				continue
			}
			dIsCmb := d[i].au.isCombination
			dau := auToString(d[i].au, false)
			dcmb := d[i].au.comb
			if isCmb == dIsCmb && au == dau && (dcmb == "" || dcmb == cmb) {
				v.data = append(v.data, d[i])
			}
		}
	}
	return gs
}

func setAuthorship(g group) {
	var auInt int
	var auLen int

	for _, v := range g.data {
		if v.p.Authorship == nil {
			continue
		}
		var vInt int
		if v.au.orig != "" {
			vInt++
		}
		if v.au.comb != "" {
			vInt++
		}
		vLen := len(v.p.Authorship.Normalized)
		if auInt < vInt && auLen < vLen {
			auInt = vInt
			auLen = vLen
			g.au = v.au
		}
	}
}

func toLexicalGroups(gs []group) []LexicalGroup {
	// sort within groups according to provided authorship and then
	// by the position in matching results.
	for i := range gs {
		slices.SortFunc(gs[i].data, func(a, b record) int {
			// results with authors are better than ones without authors.
			if a.au != nil && b.au == nil {
				return -1
			} else if a.au == nil && b.au != nil {
				return 1
			}
			// the smaller is the index (position in results), the better.
			return cmp.Compare(a.idx, b.idx)
		})
	}
	// then sort groups themselves. Take the best record from each group and
	// sort by its index. The closer record to the best result (smaller its
	// index), the better
	slices.SortFunc(gs, func(a, b group) int {
		return cmp.Compare(a.data[0].idx, b.data[0].idx)
	})

	res := make([]LexicalGroup, len(gs))
	for i := range gs {
		var lg LexicalGroup
		for j := range gs[i].data {
			if j == 0 {
				lg = New(gs[i].data[j].rd)
			} else {
				code := getCode(gs[i].data[j].rd)
				if code != "" {
					lg.NomCodes[code] = struct{}{}
				}
			}

			lg.Data = append(lg.Data, gs[i].data[j].rd)
		}
		res[i] = lg
	}
	return res
}

func getCode(rd *verifier.ResultData) string {
	if rd.MatchType == verifier.Virus {
		return "ICVCN"
	}

	if rd.ClassificationPath == "" {
		return ""
	}

	if strings.Index(rd.ClassificationPath, "Animalia") > -1 {
		return "ICZN"
	}

	if strings.Index(rd.ClassificationPath, "Plantae") > -1 {
		return "ICN"
	}

	if strings.Index(rd.ClassificationPath, "Fungi") > -1 {
		return "ICN"
	}

	if strings.Index(rd.ClassificationPath, "Bacteria") > -1 {
		return "ICN"
	}

	return ""
}
