package score

import (
	"fmt"

	vlib "github.com/gnames/gnlib/domain/entity/verifier"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Score", func() {
	Describe("String", func() {
		It("Creates string representation of the score", func() {
			s := Score{}
			Expect(s.String()).To(Equal("00000000_00000000_00000000_00000000"))
		})
	})

	DescribeTable("accepted",
		func(record_id, accepted_record_id, expected string) {
			s := Score{}
			Expect(s.accepted(record_id, accepted_record_id).String()).To(Equal(expected))
		},
		Entry("synonym", "123", "234", "00000000_00000000_00000000_00000000"),
		Entry("accepted1", "123", "123", "00000000_01000000_00000000_00000000"),
		Entry("accepted2", "123", "", "00000000_01000000_00000000_00000000"),
	)

	DescribeTable("fuzzy",
		func(edit_dist int, expected string) {
			s := Score{}
			Expect(s.fuzzy(edit_dist).String()).To(Equal(expected))
		},
		Entry("exact", 0, "00110000_00000000_00000000_00000000"),
		Entry("fuzzy1", 1, "00100000_00000000_00000000_00000000"),
		Entry("fuzzy2", 2, "00010000_00000000_00000000_00000000"),
		Entry("fuzzy3", 3, "00000000_00000000_00000000_00000000"),
		Entry("fuzzy4", 13, "00000000_00000000_00000000_00000000"),
	)

	DescribeTable("rank",
		func(can1, can2 string, card1, card2 int, expected string) {
			s := Score{}
			Expect(s.rank(can1, can2, card1, card2).String()).To(Equal(expected))
		},
		Entry("partial", "Aus bus var. cus", "Aus bus", 3, 2, "01000000_00000000_00000000_00000000"),
		Entry("binomial", "Aus bus", "Aus bus", 2, 2, "01000000_00000000_00000000_00000000"),
		Entry("exact", "Aus bus var. cus", "Aus bus var. cus", 3, 3, "10000000_00000000_00000000_00000000"),
		Entry("no match", "Aus bus var. cus", "Aus bus f. cus", 3, 3, "00000000_00000000_00000000_00000000"),
		Entry("n/a", "Aus bus cus", "Aus bus f. cus", 3, 3, "01000000_00000000_00000000_00000000"),
		Entry("n/a", "Aus bus f. cus", "Aus bus cus", 3, 3, "01000000_00000000_00000000_00000000"),
	)

	DescribeTable("curation",
		func(dsID int, curLev vlib.CurationLevel, expected string) {
			s := Score{}
			Expect(s.curation(dsID, curLev).String()).To(Equal(expected))
		},
		Entry("no cur", 67, vlib.NotCurated, "00000000_00000000_00000000_00000000"),
		Entry("auto cur", 67, vlib.AutoCurated, "00000100_00000000_00000000_00000000"),
		Entry("cur", 67, vlib.Curated, "00001000_00000000_00000000_00000000"),
		Entry("CoL", 1, vlib.Curated, "00001100_00000000_00000000_00000000"),
	)

	DescribeTable("auth",
		func(auth1, auth2 []string, year1, year2 int, expected string) {
			s := Score{}
			Expect(s.auth(auth1, auth2, year1, year2).String()).To(Equal(expected))
		},
		Entry("empty1", []string{}, []string{}, 0, 0, "00000010_00000000_00000000_00000000"),
		Entry("empty2", []string{"L."}, []string{}, 1758, 0, "00000010_00000000_00000000_00000000"),
		Entry("empty3", []string{}, []string{"L."}, 0, 1758, "00000010_10000000_00000000_00000000"),
		Entry("no match1", []string{"Banks"}, []string{"L."}, 0, 0, "00000000_00000000_00000000_00000000"),
		Entry("no match2", []string{"L."}, []string{"Banks"}, 1758, 1758, "00000000_00000000_00000000_00000000"),
		Entry("overlap", []string{"Tomm.", "L.", "Banks", "Muetze"}, []string{"Kuntze", "Linn", "Hopkins"}, 1758, 1758, "00000000_10000000_00000000_00000000"),
		Entry("full subset, yes yr", []string{"Hopkins", "L.", "Thomson"}, []string{"Thomson", "Linn."}, 1758, 1758, "00000011_00000000_00000000_00000000"),
		Entry("full subset, aprx yr1", []string{"Hopkins", "L.", "Thomson"}, []string{"Thomson", "Linn."}, 1757, 1758, "00000010_10000000_00000000_00000000"),
		Entry("full subset, aprx yr2", []string{"L.", "Thomson"}, []string{"Thomson", "Linn.", "Hopkins"}, 1757, 1756, "00000010_10000000_00000000_00000000"),
		Entry("full subset, n/a yr1", []string{"L.", "Thomson"}, []string{"Thomson", "Linn.", "Hopkins"}, 0, 1756, "00000010_00000000_00000000_00000000"),
		Entry("full subset, n/a yr2", []string{"L.", "Thomson"}, []string{"Thomson", "Linn.", "Hopkins"}, 1756, 0, "00000010_00000000_00000000_00000000"),
		Entry("full subset, no yr", []string{"L.", "Thomson"}, []string{"Thomson", "Linn.", "Hopkins"}, 1756, 1800, "00000001_00000000_00000000_00000000"),
		Entry("match, yes yr", []string{"L.", "Thomson"}, []string{"Linn", "Thomson"}, 1800, 1800, "00000011_10000000_00000000_00000000"),
		Entry("match, aprx yr", []string{"Herenson", "Thomson"}, []string{"Thomson", "H."}, 1799, 1800, "00000011_00000000_00000000_00000000"),
		Entry("match, n/a yr", []string{"Herenson", "Thomson"}, []string{"Thomson", "H."}, 0, 0, "00000010_10000000_00000000_00000000"),
		Entry("match, bad yr", []string{"Herenson", "Thomson"}, []string{"Thomson", "H."}, 1750, 1755, "00000001_10000000_00000000_00000000"),
	)

	DescribeTable("compareAuth",
		func(au1, au2 string, expected string) {
			match, giveup := compareAuth(au1, au2)
			res := fmt.Sprintf("%v|%v", match, giveup)
			Expect(res).To(Equal(expected))
		},
		Entry("no match2", "L", "Banks", "false|false"),
		Entry("no match2", "Banks", "L", "false|true"),
		Entry("no match2", "Banks", "B", "true|false"),
		Entry("no match2", "Banks", "Banz", "false|true"),
		Entry("no match2", "Banks", "Banks", "true|false"),
	)

	DescribeTable("authNormalize",
		func(auth string, expected string) {
			Expect(authNormalize(auth)).To(Equal(expected))
		},
		Entry("empty", "", ""),
		Entry("abbr1", "L.", "L"),
		Entry("abbr2", "Linn.", "Linn"),
		Entry("initial1", "A. Linn.", "Linn"),
		Entry("initial2", "A. B. Lin", "Lin"),
		Entry("initial3", "A. B.", ""),
		Entry("two words", "A. B. Koza Koza", "Koza Koza"),
	)
})
