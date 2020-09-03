package rest_test

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/gnames/gnames/encode"
	"github.com/gnames/gnames/model"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

const url = "http://:8888/"

var _ = Describe("Rest", func() {
	Describe("Ping()", func() {
		It("Gets pong from REST server", func() {
			resp, err := http.Get(url + "ping")
			Expect(err).To(BeNil())

			response, err := ioutil.ReadAll(resp.Body)
			Expect(err).To(BeNil())

			Expect(string(response)).To(Equal("pong"))
		})
	})

	Describe("GetVersion()", func() {
		It("Gets Version from REST server", func() {
			resp, err := http.Get(url + "version")
			Expect(err).To(BeNil())
			respBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).To(BeNil())

			enc := encode.GNjson{}
			var response model.Version
			err = enc.Decode(respBytes, &response)
			Expect(err).To(BeNil())
			Expect(response.Version).To(MatchRegexp(`^v\d+\.\d+\.\d+`))
		})
	})

	Describe("Verify()", func() {
		It("Verifies entered names", func() {
			var response []model.Verification
			names := []string{
				"Not name", "Bubo bubo", "Pomatomus",
				"Pardosa moesta", "Plantago major var major",
				"Cytospora ribis mitovirus 2",
				"A-shaped rods", "Alb. alba",
			}
			request := model.VerifyParams{NameStrings: names}
			req, err := encode.GNjson{}.Encode(request)
			Expect(err).To(BeNil())
			r := bytes.NewReader(req)
			resp, err := http.Post(url+"verification", "application/x-binary", r)
			Expect(err).To(BeNil())
			respBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).To(BeNil())

			err = encode.GNjson{}.Decode(respBytes, &response)
			Expect(err).To(BeNil())
			Expect(len(response)).To(Equal(len(names)))
		})
	})
})
