package xiami

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"testing"
)

var _ = BeforeSuite(func() {
})

var _ = AfterSuite(func() {
})

func TestWorkspace(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Workspace Suite")
}
