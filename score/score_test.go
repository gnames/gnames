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
		func(uuid1, uuid2, expected string) {
			s := Score{}
			Expect(s.rank(uuid1, uuid2).String()).To(Equal(expected))
		},
		Entry("empty 1", "", "", "01000000000000000000000000000000: 1073741824"),
		Entry("empty 2", "", "123", "01000000000000000000000000000000: 1073741824"),
		Entry("yes", "123", "123", "11000000000000000000000000000000: 3221225472"),
		Entry("no", "123", "1234", "00000000000000000000000000000000: 0"),
	)
})
