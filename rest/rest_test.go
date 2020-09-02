package rest_test

import (
	"io/ioutil"
	"net/http"

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
})
