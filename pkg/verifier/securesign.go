package verifier

import (
	"context"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	securesignGVK = schema.GroupVersionKind{
		Group:   "rhtas.redhat.com",
		Version: "v1alpha1",
		Kind:    "Securesign",
	}
)

// Get retrieves a Securesign CR instance
func Get(ctx context.Context, cli client.Client, namespace, name string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(securesignGVK)

	err := cli.Get(ctx, client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}, obj)

	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return nil
	}
	return obj
}

// IsReady checks if the Securesign CR has Ready condition set to True
func IsReady(obj *unstructured.Unstructured) bool {
	if obj == nil {
		return false
	}

	conditions, found, err := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if !found || err != nil {
		return false
	}

	for _, cond := range conditions {
		condMap, ok := cond.(map[string]interface{})
		if !ok {
			continue
		}

		condType, ok := condMap["type"].(string)
		if !ok || condType != "Ready" {
			continue
		}

		status, ok := condMap["status"].(string)
		return ok && status == "True"
	}

	return false
}

// Verify waits for the Securesign CR to be ready
func Verify(ctx context.Context, cli client.Client, namespace, name string) {
	Eventually(func(g Gomega) *unstructured.Unstructured {
		obj := Get(ctx, cli, namespace, name)
		g.Expect(obj).NotTo(BeNil())
		return obj
	}).WithContext(ctx).Should(Not(BeNil()))

	Eventually(func(g Gomega) bool {
		obj := Get(ctx, cli, namespace, name)
		g.Expect(obj).NotTo(BeNil())
		return IsReady(obj)
	}).WithContext(ctx).Should(BeTrue())
}

// GetReadyCondition is a Gomega matcher helper
func GetReadyCondition(obj *unstructured.Unstructured) types.GomegaMatcher {
	return WithTransform(func(o *unstructured.Unstructured) bool {
		return IsReady(o)
	}, BeTrue())
}

