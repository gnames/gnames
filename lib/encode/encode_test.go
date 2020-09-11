package encode_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/gnames/gnames/lib/encode"
)

type version struct {
	Version string
	Build   string
}

var _ = Describe("Encode", func() {
	Describe("Encoder", func() {
		It("encodes and decodes a string", func() {
			encs := []Encoder{
				GNgob{},
				GNjson{},
			}
			for _, e := range encs {
				obj := version{
					Version: "v10.10.10",
					Build:   "today",
				}
				res, err := e.Encode(obj)
				Expect(err).To(BeNil())
				var ver version
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
			obj := version{
				Version: "v10.10.10",
				Build:   "today",
			}
			res, err := enc.Encode(obj)
			Expect(err).To(BeNil())
			var ver version
			err = enc.Decode(res, &ver)
			Expect(err).To(BeNil())
			Expect(ver.Version).To(Equal("v10.10.10"))
			Expect(ver.Build).To(Equal("today"))
		})
	})

	Describe("GNjson", func() {
		It("encodes and decodes a string", func() {
			enc := GNjson{}
			obj := version{
				Version: "v10.10.10",
				Build:   "today",
			}
			res, err := enc.Encode(obj)
			Expect(err).To(BeNil())
			var ver version
			err = enc.Decode(res, &ver)
			Expect(err).To(BeNil())
			Expect(ver.Version).To(Equal("v10.10.10"))
			Expect(ver.Build).To(Equal("today"))
		})
	})
})

func ExampleEncodeDecode() {
	var enc Encoder
	var err error
	enc = GNjson{Pretty: true}
	ary1 := []int{1, 2, 3}
	jsonRes, err := enc.Encode(ary1)
	if err != nil {
		panic(err)
	}
	var ary2 []int
	err = enc.Decode(jsonRes, &ary2)
	if err != nil {
		panic(err)
	}
	fmt.Println(ary1[0] == ary2[0])
	// Output: true
}
