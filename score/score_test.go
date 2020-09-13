package score

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Score", func() {
	Describe("String", func() {
		It("Creates string representation of the score", func() {
			s := Score{}
			Expect(s.String()).To(Equal("00000000000000000000000000000000: 0"))
		})
	})

	DescribeTable("rank",
		func(can1, can2 string, card1, card2 int, expected string) {
			s := Score{}
			Expect(s.rank(can1, can2, card1, card2).String()).To(Equal(expected))
		},
		Entry("partial", "Aus bus var. cus", "Aus bus", 3, 2, "01000000000000000000000000000000: 1073741824"),
		Entry("binomial", "Aus bus", "Aus bus", 2, 2, "01000000000000000000000000000000: 1073741824"),
		Entry("exact", "Aus bus var. cus", "Aus bus var. cus", 3, 3, "10000000000000000000000000000000: 2147483648"),
		Entry("no match", "Aus bus var. cus", "Aus bus f. cus", 3, 3, "00000000000000000000000000000000: 0"),
		Entry("n/a", "Aus bus cus", "Aus bus f. cus", 3, 3, "01000000000000000000000000000000: 1073741824"),
		Entry("n/a", "Aus bus f. cus", "Aus bus cus", 3, 3, "01000000000000000000000000000000: 1073741824"),
	)
})
