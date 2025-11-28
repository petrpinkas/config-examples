package support

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CreateTestNamespace creates a test namespace with a generated name based on the test file
func CreateTestNamespace(ctx ginkgo.SpecContext, cli client.Client) *v1.Namespace {
	sp := ginkgo.CurrentSpecReport()
	var name string
	if sp.LeafNodeLocation.FileName != "" {
		fn := filepath.Base(sp.LeafNodeLocation.FileName)
		// Replace invalid characters with '-'
		re := regexp.MustCompile("[^a-z0-9-]")
		name = re.ReplaceAllString(strings.TrimSuffix(fn, filepath.Ext(fn)), "-")
	} else {
		name = "rhtas-test"
	}

	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: name + "-",
		},
	}
	Expect(cli.Create(ctx, ns)).To(Succeed())
	ginkgo.GinkgoWriter.Printf("Created test namespace: %s\n", ns.Name)
	return ns
}

// CreateNamespaceStep returns a BeforeAll function that creates a namespace and registers cleanup
func CreateNamespaceStep(cli client.Client, callback func(*v1.Namespace)) func(ctx ginkgo.SpecContext) {
	return func(ctx ginkgo.SpecContext) {
		namespace := CreateTestNamespace(ctx, cli)
		ginkgo.DeferCleanup(func(ctx ginkgo.SpecContext) {
			ginkgo.GinkgoWriter.Printf("Deleting test namespace: %s\n", namespace.Name)
			Expect(cli.Delete(ctx, namespace)).To(Succeed())
		})
		callback(namespace)
	}
}
