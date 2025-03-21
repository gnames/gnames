package score

import (
	"strings"

	vlib "github.com/gnames/gnlib/ent/verifier"
)

// score calculates and stores the score of a match.
//
// score contains 32bits. They are distributed between different criteria the
// following way (from the left):
//
// x0000000_00000000_00000000_00000000: matching cardinality
// 00 - cardinality does not match
//
//	`Aus bus cus` vs `Aus bus`
//
// 01 - cardinality match
//
//	`Aus bus cus` vs `Aus bus f. cus`
//
// 0xx00000_00000000_00000000_00000000: matching rank for ranked infraspecies
// 00 - rank does not match
//
//	`Aus bus var. cus` vs `Aus bus f. cus`
//
// 01 - match is unknown
//
//	`Aus bus cus` vs `Aus bus f. cus`
//
// 10 - rank matches
//
//	`Aus bus var. cus` vs `Aus bus var. cus`
//
// 000xx000_00000000_00000000_00000000: edit distance
// 00 - edit distance is 3 or more
// 01 - edit distance 2
// 10 - edit distance 1
// 11 - edit distance 0
//
// 00000xxx_00000000_00000000_00000000: matching authorship
//
// 000 - authorship does not match.
//
//	`Linn.` vs `Banks`
//
// 001 - Authorship is not comparable.
//
//	`` vs ``
//
// 001 - Authorship is not comparable, input has no authorship, but
//
//	output has authorship.
//	`` vs `Auth1, Auth2, 1880`
//
// 010 - Authors overlap, but years do not match.
//
//	`Auth1, Auth2 1778` vs `Auth1, Auth3 1785`
//
// 011 - one set of authors fully included into other set,
//
//	yers do not match.
//	`Auth1, Auth2, 1880` vs `Auth1, Auth2, Auth3, 1887`
//
// 100 - Authors are identical, years do not match.
//
//	`Auth1, Auth2, 1880` vs `Auth1, Auth2, 1887`
//
// 101 - a,y+: Authors overlap, but there are additional authors
//
//	for both sets. Years either match somewhat or not available.
//	`Auth1, Auth2, 1888` vs `Auth1, Auth3, 1887`
//
// 101 - aaa,y?: Authors are identical, year is not comparable.
//
//	`Auth1, Auth2` vs `Auth1, Auth2, 1888`
//
// 101 - aa,yy: One set of authors is fully included into other,
//
//	years are close.
//	`Auth1, Auth2, Auth3, 1888` vs `Auth1, Auth2, 1887`
//
// 110 - aa,yyy: One set of authors is fully included into another,
//
//	same years.
//
// 110 - aaa,yy: Authors are identical, years are close.
//
//	`Auth1, Auth2, 1888` vs `Auth1, Auth2, 1887`
//	`Auth1, Auth2, Auth3, 1888` vs `Auth1, Auth2, 1888`
//
// 111 - aaa,yyy: Authors and years are identical.
//
//	`Auth1, Auth2, 1888` vs `Auth1, Auth2, 1888`
//
// 00000000_xx000000_00000000_00000000: curation
// 00 - uncurated sources only
// 01 - auto-curated sources
// 10 - human-curated sources
// 11 - Catalogue of Life
//
// 00000000_00x00000_00000000_00000000: accepted name
// 0 - name is a synonym
// 1 - name is currently accepted
//
// 00000000_000xx000_00000000_00000000: parsing quality of a match
// 00 - parsing failed
// 01 - significant parsing problems
// 10 - some parsing problems
// 11 - clean parsing
type score struct {
	value uint32
}

var (
	cardinalityShift    = 31
	rankShift           = 29
	fuzzyShift          = 27
	authShift           = 24
	curationShift       = 22
	acceptedShift       = 21
	parsingQualityShift = 19
)

// cardinality checks if canonical forms have the same cardinality.
// If cardinality matches, result is 1, if not, result is 0. If
// cardinality is unknown, result is 0.
func (s score) cardinality(card1, card2 int) score {
	if card1 == 0 || card2 == 0 || card1 != card2 {
		return s
	}
	s.value = s.value | 0b01<<cardinalityShift
	return s
}

func (s score) cardinalityVal() float32 {
	return s.extractVal(cardinalityShift, 0b01, 1)
}

// rank checks if infraspecific canonical forms contain the same ranks. If they
// do the score is 2, if comparison cannot be done the score is 1, and if the
// ranks are different, the score is 0. 2 bits, shift 30
func (s score) rank(can1, can2 string, card1, card2 int) score {
	if card1 < 3 || card1 != card2 {
		return s
	}

	ranks1 := getRanks(can1)
	ranks2 := getRanks(can2)

	if ranks1 == "" || ranks2 == "" {
		s.value = s.value | 0b01<<rankShift
		return s
	}

	if ranks1 == ranks2 {
		s.value = s.value | 0b10<<rankShift
	}

	return s
}

func getRanks(s string) string {
	words := strings.Split(s, " ")

	// should never happen
	if len(words) < 3 {
		return ""
	}

	words = words[2:]

	var res string
	for i := range words {
		if strings.HasSuffix(words[i], ".") {
			res += words[i]
		}
	}
	return res
}

func (s score) rankVal() float32 {
	return s.extractVal(rankShift, 0b11, 2)
}

// fuzzy matching
func (s score) fuzzy(editDistance int) score {
	ed := editDistance
	var i uint32 = 3
	if ed > int(i) || ed < 0 {
		i = 0
	} else {
		i = i - uint32(ed)
	}

	s.value = s.value | i<<fuzzyShift
	return s
}

func (s score) fuzzyVal() float32 {
	val := s.extractVal(fuzzyShift, 0b11, 3)
	return val
}

// curation scores by curation level of data-sources.
func (s score) curation(dataSourceID int,
	curationLevel vlib.CurationLevel) score {
	var i uint32

	if dataSourceID == 1 {
		s.value = s.value | 0b11<<curationShift
		return s
	}

	switch curationLevel {
	case vlib.NotCurated:
		i = 0b00
	case vlib.AutoCurated:
		i = 0b01
	case vlib.Curated:
		i = 0b10
	}
	s.value = s.value | i<<curationShift
	return s
}

func (s score) curationVal() float32 {
	return s.extractVal(curationShift, 0b11, 3)
}

// auth takes two lists of authors, and their corresponding years and
// tries to match them to each other. The score is decided on how well
// authors and years did match.
// The score takes 3 bits and ranges from 0 to 7.
func (s score) auth(auth1, auth2 []string, year1, year2 int) score {
	years := findYearsMatch(year1, year2)
	authors := findAuthMatch(auth1, auth2)
	var i uint32 = 0

	if authors == identical {
		switch years {
		case perfectMatch:
			i = 0b111 //7
		case approxMatch:
			i = 0b110 //6
		case notAvailable:
			i = 0b101 //5
		case noMatch:
			i = 0b001 //1
		}
	} else if authors == fullInclusion {
		switch years {
		case perfectMatch:
			i = 0b110 //6
		case approxMatch:
			i = 0b101 //5
		case notAvailable:
			i = 0b100 //4
		case noMatch:
			i = 0b001 //1
		}
	} else if authors == overlap {
		switch years {
		case perfectMatch:
			i = 0b101 //5
		case approxMatch:
			i = 0b100 //4
		case notAvailable:
			i = 0b010 //2
		case noMatch:
			i = 0b001 //1
		}
	} else if authors == noAuthVsAuth {
		// decreased it to 1 from 2, to prevent synonym match first like in case
		// `Diptera` match with `Diptera Borkh.` (synonym of `Saxifraga L.`)
		i = 0b001 //1
	} else if authors == incomparable {
		i = 0b001 //1
	}
	// authors do not match so by default:
	// i = 0b00 //

	s.value = s.value | i<<authShift
	return s
}

func (s score) authVal() float32 {
	return s.extractVal(authShift, 0b111, 7)
}

// accepted name
func (s score) accepted(recordID, acceptedID string) score {
	var i uint32 = 0
	if acceptedID == "" || recordID == acceptedID {
		i = 1
	}
	s.value = s.value | i<<acceptedShift
	return s
}

func (s score) acceptedVal() float32 {
	return s.extractVal(acceptedShift, 0b1, 1)
}

// parsingQuality
func (s score) parsingQuality(quality int) score {
	var i uint32 = 0
	switch quality {
	case 3:
		i = 0b01
	case 2:
		i = 0b10
	case 1:
		i = 0b11
	}
	s.value = s.value | i<<parsingQualityShift
	return s
}

func (s score) parsingQualityVal() float32 {
	return s.extractVal(parsingQualityShift, 0b11, 3)
}

func (s score) extractVal(shift, mask int, max float32) float32 {
	masked := s.value & uint32(mask<<shift)
	res := float32(masked>>shift) / max
	return res
}
