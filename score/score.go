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
// 11 - rank matches
//    `Aus bus var. cus` vs `Aus bus var. cus`
//
package score

import (
	"fmt"
)

type Score struct {
	Value uint32
}

func (s Score) String() string {
	return fmt.Sprintf("%032b: %d", s.Value, s.Value)
}

func (s *Score) rank(uuid1, uuid2 string) *Score {
	shift := 30
	i := s.Value

	if uuid1 == "" || uuid2 == "" {
		s.Value = (i | uint32(0b01<<shift))
		return s
	}

	if uuid1 == uuid2 {
		s.Value = (i | uint32(0b11<<shift))
	}

	return s
}
