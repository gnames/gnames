// Package score provides a single number that allows to sort matching results
// according to their relevancy.
// Score contains 32bits. They are distributed between different criteria the
// following way (from the left):

// xx000000_00000000_00000000_00000000: matching rank for ranked infraspecies
// 00 - rank does not match
//    `Aus bus var. cus` vs `Aus bus f. cus`
// 01 - match is unknown
//    `Aus bus cus` vs `Aus bus f. cus`
//    `Aus bus` vs `Aus bus`
// 10 - rank matches
//    `Aus bus var. cus` vs `Aus bus var. cus`
//
// 00xx0000_00000000_00000000_00000000: curation
// 00 - uncurated sources only
// 01 - auto-curated sources
// 10 - human-curated sources
// 11 - Catalogue of Life
//
// 0000xxx0_00000000_00000000_00000000: matching authorship
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
// 0000000x_00000000_00000000_00000000: accepted name
// 0 - name is a synonym
// 1 - name is currently accepted
//
// 00000000_xx000000_00000000_00000000: edit distance
// 00 - edit distance is 3 or more
// 01 - edit distance 2
// 10 - edit distance 1
// 11 - edit distance 0
package score

import (
	"fmt"
	"strings"

	"github.com/gnames/gnames/domain/entity"
)

// Score
type Score struct {
	Value uint32
}

// String returns a string representation of a score as a set of bits with
// every byte (8 bits) separated by an underscore.
func (s Score) String() string {
	str := fmt.Sprintf("%032b", s.Value)
	res := make([]byte, 35)
	offset := 0
	for i, v := range []byte(str) {
		res[i+offset] = v
		if (i+1)%8 == 0 && (i+1)%32 != 0 {
			offset++
			res[i+offset] = '_'
		}
	}
	return string(res)
}

// rank checks if infraspecific canonical forms contain the same ranks. If they
// do the score is 2, if comparison cannot be done the score is 1, and if the
// ranks are different, the score is 0. 2 bits, shift 30
func (s Score) rank(can1, can2 string, card1, card2 int) Score {
	shift := 30
	i := s.Value

	if card1 < 3 || card1 != card2 ||
		!strings.Contains(can1, ".") || !strings.Contains(can2, ".") {
		s.Value = (i | uint32(0b01<<shift))
		return s
	}

	if can1 == can2 {
		s.Value = (i | uint32(0b10<<shift))
	}
	return s
}

// curation scores by curation level of data-sources.
func (s Score) curation(dataSourceID int, curationLevel entity.CurationLevel) Score {
	shift := 28
	i := uint32(curationLevel)
	if dataSourceID == 1 {
		i = 3
	}
	s.Value = (s.Value | uint32(i<<shift))
	return s
}

// accepted name
func (s Score) accepted(record_id, accepted_id string) Score {
	shift := 24
	var i uint32 = 0
	if accepted_id == "" || record_id == accepted_id {
		i = 1
	}
	s.Value = (s.Value | uint32(i<<shift))
	return s
}

// fuzzy matching
func (s Score) fuzzy(edit_distance int) Score {
	shift := 22
	var i uint32 = 3
	if edit_distance > int(i) || edit_distance < 0 {
		i = 0
	} else {
		i = i - uint32(edit_distance)
	}

	s.Value = (s.Value | uint32(i<<shift))
	return s
}
