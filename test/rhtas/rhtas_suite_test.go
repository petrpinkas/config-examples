package rhtas

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRhtas(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "RHTAS Configuration Tests")
}
