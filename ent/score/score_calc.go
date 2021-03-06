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
// xx000000_00000000_00000000_00000000: matching rank for ranked infraspecies
// 00 - rank does not match
//    `Aus bus var. cus` vs `Aus bus f. cus`
// 01 - match is unknown
//    `Aus bus cus` vs `Aus bus f. cus`
//    `Aus bus` vs `Aus bus`
// 10 - rank matches
//    `Aus bus var. cus` vs `Aus bus var. cus`
//
// 00xx0000_00000000_00000000_00000000: edit distance
// 00 - edit distance is 3 or more
// 01 - edit distance 2
// 10 - edit distance 1
// 11 - edit distance 0
//
// 0000xx00_00000000_00000000_00000000: curation
// 00 - uncurated sources only
// 01 - auto-curated sources
// 10 - human-curated sources
// 11 - Catalogue of Life
//
// 000000xx_x0000000_00000000_00000000: matching authorship
// 000 - authorship does not match
//       `Linn.` vs `Banks`
// 001 - authors overlap, but there are additional authors for both sets. Years
//       either match somewhat or not available.
//       `Auth1, Auth2, 1888` vs `Auth1, Auth3, 1887`
// 010 - one set of authors fully included into other set,
//       yers do not match.
//       `Auth1, Auth2, 1880` vs `Auth1, Auth2, Auth3, 1887`
// 011 - Authors are identical, years do not match.
//       `Auth1, Auth2, 1880` vs `Auth1, Auth2, 1887`
// 100 - Authorship is not comparable.
//       `Auth1, Auth2, 1880` vs ``
// 101 - Authorship is not comparable, input has no authorship, but
//       output has authorship
// 101 - Authors are identical, year is not comparable.
//       `Auth1, Auth2` vs `Auth1, Auth2, 1888`
// 101 - One set of authors is fully included into other,
//       years are close.
//       `Auth1, Auth2, Auth3, 1888` vs `Auth1, Auth2, 1887`
// 110 - Authors are identical, years are close.
//       `Auth1, Auth2, 1888` vs `Auth1, Auth2, 1887`
// 110 - One set of authors is fully included into another,
//       same years.
//       `Auth1, Auth2, Auth3, 1888` vs `Auth1, Auth2, 1888`
// 111 - Authors and years are identical.
//       `Auth1, Auth2, 1888` vs `Auth1, Auth2, 1888`
//
// 00000000_0x000000_00000000_00000000: accepted name
// 0 - name is a synonym
// 1 - name is currently accepted
//
// 00000000_00xx0000_00000000_00000000: parsing quality of a match
// 00 - parsing failed
// 01 - significant parsing problems
// 10 - some parsing problems
// 11 - clean parsing
type score struct {
	value uint32
}

// rank checks if infraspecific canonical forms contain the same ranks. If they
// do the score is 2, if comparison cannot be done the score is 1, and if the
// ranks are different, the score is 0. 2 bits, shift 30
func (s score) rank(can1, can2 string, card1, card2 int) score {
	shift := 30
	if card1 < 3 || card1 != card2 ||
		!strings.Contains(can1, ".") || !strings.Contains(can2, ".") {
		s.value = (s.value | uint32(0b01<<shift))
		return s
	}

	if can1 == can2 {
		s.value = (s.value | uint32(0b10<<shift))
	}
	return s
}

// fuzzy matching
func (s score) fuzzy(editDistance int) score {
	ed := editDistance
	shift := 28
	var i uint32 = 3
	if ed > int(i) || ed < 0 {
		i = 0
	} else {
		i = i - uint32(ed)
	}

	s.value = (s.value | uint32(i<<shift))
	return s
}

// curation scores by curation level of data-sources.
func (s score) curation(dataSourceID int,
	curationLevel vlib.CurationLevel) score {
	shift := 26
	var i uint32

	if dataSourceID == 1 {
		s.value = (s.value | 0b11<<shift)
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
	s.value = (s.value | i<<shift)
	return s
}

// auth takes two lists of authors, and their corresponding years and
// tries to match them to each other. The score is decided on how well
// authors and years did match.
// The score takes 3 bits and ranges from 0 to 7.
func (s score) auth(auth1, auth2 []string, year1, year2 int) score {
	shift := 23
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
			i = 0b011 //3
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
			i = 0b010 //2
		}
	} else if authors == overlap {
		switch years {
		case noMatch:
			i = 0b000 //0
		default:
			i = 0b001 //1
		}
	} else if authors == noAuthVsAuth {
		i = 0b101 //5
	} else if authors == uncomparable {
		i = 0b100 //4
	}
	s.value = (s.value | i<<shift)
	return s
}

// accepted name
func (s score) accepted(recordID, acceptedID string) score {
	shift := 22
	var i uint32 = 0
	if acceptedID == "" || recordID == acceptedID {
		i = 1
	}
	s.value = (s.value | uint32(i<<shift))
	return s
}

// parsingQuality
func (s score) parsingQuality(quality int) score {
	shift := 20
	var i uint32 = 0
	switch quality {
	case 3:
		i = 0b01
	case 2:
		i = 0b10
	case 1:
		i = 0b11
	}
	s.value = (s.value | uint32(i<<shift))
	return s
}
