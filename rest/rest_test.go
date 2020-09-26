package rest_test

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/gnames/gnames/domain/entity"
	"github.com/gnames/gnames/lib/encode"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	// log "github.com/sirupsen/logrus"
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

	Describe("Version()", func() {
		It("Gets Version from REST server", func() {
			resp, err := http.Get(url + "version")
			Expect(err).To(BeNil())
			respBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).To(BeNil())

			enc := encode.GNjson{}
			var response entity.Version
			err = enc.Decode(respBytes, &response)
			Expect(err).To(BeNil())
			Expect(response.Version).To(MatchRegexp(`^v\d+\.\d+\.\d+`))
		})
	})

	Describe("Verify()", func() {
		It("Verifies entered names", func() {
			var response []entity.Verification
			names := []string{
				"Not name", "Bubo bubo", "Pomatomus",
				"Pardosa moesta", "Plantago major var major",
				"Cytospora ribis mitovirus 2",
				"A-shaped rods", "Alb. alba",
				"Pisonia grandis", "Acacia vestita may",
			}
			request := entity.VerifyParams{NameStrings: names}
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

			bad := response[0]
			Expect(bad.InputID).To(Equal("82dbfb99-fe6c-5882-99f2-17c7d3955599"))
			Expect(bad.Input).To(Equal("Not name"))
			Expect(bad.MatchType).To(Equal(entity.NoMatch))
			Expect(bad.BestResult).To(BeNil())
			Expect(bad.DataSourcesNum).To(Equal(0))
			Expect(bad.CurationLevel).To(Equal(entity.NotCurated))
			Expect(bad.Error).To(Equal(""))

			binom := response[1]
			Expect(binom.InputID).To(Equal("4431a0f3-e901-519a-886f-9b97e0c99d8e"))
			Expect(binom.Input).To(Equal("Bubo bubo"))
			Expect(binom.BestResult).ToNot(BeNil())
			Expect(binom.BestResult.DataSourceID).To(Equal(1))
			Expect(binom.BestResult.MatchType).To(Equal(entity.Exact))
			Expect(binom.CurationLevelString).To(Equal("CURATED"))
			Expect(binom.Error).To(Equal(""))

			acceptFilter := response[8]
			Expect(acceptFilter.InputID).To(Equal("4c8848f2-7271-588c-ba81-e4d5efcc1e92"))
			Expect(acceptFilter.Input).To(Equal("Pisonia grandis"))
			Expect(acceptFilter.BestResult.DataSourceID).To(Equal(1))
			Expect(acceptFilter.BestResult.MatchType).To(Equal(entity.Exact))
			Expect(acceptFilter.BestResult.CurrentCanonicalSimple).To(Equal("Pisonia grandis"))

			partial := response[9]
			Expect(partial.InputID).To(Equal("0f84ed48-3a57-59ac-ac1a-2e9221439fdc"))
			Expect(partial.Input).To(Equal("Acacia vestita may"))
			Expect(partial.BestResult.DataSourceID).To(Equal(1))
			Expect(partial.MatchType).To(Equal(entity.PartialExact))
			Expect(partial.BestResult.CurrentCanonicalSimple).To(Equal("Acacia vestita"))
		})

		It("does fuzzy verification correctly", func() {
			var response []entity.Verification
			names := []string{
				"Abras precatorius",
			}
			request := entity.VerifyParams{NameStrings: names, PreferredSources: []int{1, 12, 169, 182}}
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

			fuz1 := response[0]
			Expect(fuz1.Input).To(Equal("Abras precatorius"))
			Expect(fuz1.BestResult.EditDistance).To(Equal(1))
		})

		It("Verifies entered names with preferred data-sources", func() {
			var response []entity.Verification
			names := []string{
				"Bubo bubo", "Pomatomus",
				"Pardosa moesta", "Plantago major var major",
				"Cytospora ribis mitovirus 2",
				"Pisonia grandis",
			}
			request := entity.VerifyParams{NameStrings: names, PreferredSources: []int{1, 12, 169, 182}}
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

			binom := response[0]
			Expect(binom.InputID).To(Equal("4431a0f3-e901-519a-886f-9b97e0c99d8e"))
			Expect(binom.Input).To(Equal("Bubo bubo"))
			Expect(binom.BestResult).ToNot(BeNil())
			Expect(binom.BestResult.DataSourceID).To(Equal(1))
			Expect(binom.BestResult.MatchType).To(Equal(entity.Exact))
			Expect(binom.CurationLevelString).To(Equal("CURATED"))
			Expect(len(binom.PreferredResults)).To(Equal(3))
			Expect(binom.Error).To(Equal(""))

			acceptFilter := response[5]
			Expect(acceptFilter.InputID).To(Equal("4c8848f2-7271-588c-ba81-e4d5efcc1e92"))
			Expect(acceptFilter.Input).To(Equal("Pisonia grandis"))
			Expect(acceptFilter.BestResult.DataSourceID).To(Equal(1))
			Expect(acceptFilter.BestResult.MatchType).To(Equal(entity.Exact))
			Expect(acceptFilter.BestResult.CurrentCanonicalSimple).To(Equal("Pisonia grandis"))
			Expect(len(binom.PreferredResults)).To(Equal(3))
		})

		It("Verifies names that were breaking older versions", func() {
			var response []entity.Verification
			names := []string{
				"Aceratagallia fuscosscripta (Oman )",
				"Ampullaria immersa",
				"Abacetine",
			}
			request := entity.VerifyParams{NameStrings: names}
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

	Describe("DataSources()", func() {
		It("Receives metadata of all data sources", func() {
			var response []*entity.DataSource
			req := []byte("")
			r := bytes.NewReader(req)
			resp, err := http.Post(url+"data_sources", "application/x-binary", r)
			Expect(err).To(BeNil())
			respBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).To(BeNil())

			err = encode.GNjson{}.Decode(respBytes, &response)
			Expect(err).To(BeNil())
			Expect(len(response)).To(BeNumerically(">", 50))
			col := response[0]
			Expect(col.Title).To(Equal("Catalogue of Life"))
		})

		It("Receives metadata of a data source", func() {
			var response []*entity.DataSource
			req := []byte("")
			r := bytes.NewReader(req)
			resp, err := http.Post(url+"data_sources/12", "application/x-binary", r)
			Expect(err).To(BeNil())
			respBytes, err := ioutil.ReadAll(resp.Body)
			Expect(err).To(BeNil())

			err = encode.GNjson{}.Decode(respBytes, &response)
			Expect(err).To(BeNil())
			Expect(len(response)).To(Equal(1))
			col := response[0]
			Expect(col.Title).To(Equal("Encyclopedia of Life"))
		})
	})
})
