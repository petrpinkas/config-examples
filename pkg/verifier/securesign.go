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

// Get retrieves a resource instance by GroupVersionKind
func Get(ctx context.Context, cli client.Client, namespace, name string, gvk schema.GroupVersionKind) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)

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

// GetSecuresign retrieves a Securesign CR instance (backward compatibility)
func GetSecuresign(ctx context.Context, cli client.Client, namespace, name string) *unstructured.Unstructured {
	return Get(ctx, cli, namespace, name, securesignGVK)
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

// Verify waits for a resource to be ready
func Verify(ctx context.Context, cli client.Client, namespace, name string, gvk schema.GroupVersionKind) {
	Eventually(func(g Gomega) *unstructured.Unstructured {
		obj := Get(ctx, cli, namespace, name, gvk)
		g.Expect(obj).NotTo(BeNil())
		return obj
	}).WithContext(ctx).Should(Not(BeNil()))

	Eventually(func(g Gomega) bool {
		obj := Get(ctx, cli, namespace, name, gvk)
		g.Expect(obj).NotTo(BeNil())
		return IsReady(obj)
	}).WithContext(ctx).Should(BeTrue())
}

// VerifySecuresign waits for the Securesign CR to be ready (backward compatibility)
func VerifySecuresign(ctx context.Context, cli client.Client, namespace, name string) {
	Verify(ctx, cli, namespace, name, securesignGVK)
}

// GetReadyCondition is a Gomega matcher helper
func GetReadyCondition(obj *unstructured.Unstructured) types.GomegaMatcher {
	return WithTransform(func(o *unstructured.Unstructured) bool {
		return IsReady(o)
	}, BeTrue())
}
