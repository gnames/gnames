package score

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestScore(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Score Suite")
}
