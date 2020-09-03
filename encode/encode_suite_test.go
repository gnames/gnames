package encode_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEncode(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Encode Suite")
}
