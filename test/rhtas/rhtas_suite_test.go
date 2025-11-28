package rhtas

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestRhtas(t *testing.T) {
	// Setup logging
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logrus.SetLevel(logrus.InfoLevel)

	RegisterFailHandler(Fail)
	RunSpecs(t, "RHTAS Configuration Tests")
}
