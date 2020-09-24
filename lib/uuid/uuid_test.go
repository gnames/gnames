package uuid_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gnames/gnames/lib/uuid"
)

var _ = Describe("uuid", func() {
	Describe("GNDomain", func() {
		It("contains globalnames.org domain for UUID5", func() {
			Expect(GNDomain.String()).To(Equal("90181196-fecf-5082-a4c1-411d4f314cda"))
		})
	})

	Describe("Nil", func() {
		It("contains empty UUID", func() {
			Expect(Nil.String()).To(Equal("00000000-0000-0000-0000-000000000000"))
		})
	})

	Describe("New", func() {
		It("generates new UUID5 ID for globalnames", func() {
			Expect(New("Homo sapiens").String()).To(Equal("16f235a0-e4a3-529c-9b83-bd15fe722110"))
		})
	})
})
