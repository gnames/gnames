package gnames_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGnames(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gnames Suite")
}
