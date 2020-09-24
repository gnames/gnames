package uuid_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestUuid(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Uuid Suite")
}
