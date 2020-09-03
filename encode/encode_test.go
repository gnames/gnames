package encode_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gnames/gnames/encode"
	"github.com/gnames/gnames/model"
)

var _ = Describe("Encode", func() {
	Describe("Encoder", func() {
		It("encodes and decodes a string", func() {
			encs := []Encoder{
				GNgob{},
				GNjson{},
			}
			for _, e := range encs {
				obj := model.Version{
					Version: "v10.10.10",
					Build:   "today",
				}
				res, err := e.Encode(obj)
				Expect(err).To(BeNil())
				var ver model.Version
				err = e.Decode(res, &ver)
				Expect(err).To(BeNil())
				Expect(ver.Version).To(Equal("v10.10.10"))
				Expect(ver.Build).To(Equal("today"))
			}
		})
	})

	Describe("GNgob", func() {
		It("encodes and decodes a string", func() {
			enc := GNgob{}
			obj := model.Version{
				Version: "v10.10.10",
				Build:   "today",
			}
			res, err := enc.Encode(obj)
			Expect(err).To(BeNil())
			var ver model.Version
			err = enc.Decode(res, &ver)
			Expect(err).To(BeNil())
			Expect(ver.Version).To(Equal("v10.10.10"))
			Expect(ver.Build).To(Equal("today"))
		})
	})

	Describe("GNjson", func() {
		It("encodes and decodes a string", func() {
			enc := GNjson{}
			obj := model.Version{
				Version: "v10.10.10",
				Build:   "today",
			}
			res, err := enc.Encode(obj)
			Expect(err).To(BeNil())
			var ver model.Version
			err = enc.Decode(res, &ver)
			Expect(err).To(BeNil())
			Expect(ver.Version).To(Equal("v10.10.10"))
			Expect(ver.Build).To(Equal("today"))
		})
	})
})

