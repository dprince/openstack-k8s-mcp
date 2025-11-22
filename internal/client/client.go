package client

import (
	"context"
	"encoding/json"
	"fmt"

	openstackv1beta1 "github.com/openstack-k8s-operators/openstack-operator/apis/core/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	openstackVersionGVR = schema.GroupVersionResource{
		Group:    "core.openstack.org",
		Version:  "v1beta1",
		Resource: "openstackversions",
	}
)

// K8sClient wraps Kubernetes client functionality
type K8sClient struct {
	client dynamic.Interface
}

// NewK8sClient creates a new Kubernetes client
func NewK8sClient() (*K8sClient, error) {
	config, err := getKubeConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	dynClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	return &K8sClient{client: dynClient}, nil
}

// getKubeConfig attempts to get kubeconfig from in-cluster or kubeconfig file
func getKubeConfig() (*rest.Config, error) {
	// Try in-cluster config first
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Fall back to kubeconfig file
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	return kubeConfig.ClientConfig()
}

// GetOpenStackVersion retrieves OpenStackVersion CR from the specified namespace
func (c *K8sClient) GetOpenStackVersion(ctx context.Context, namespace, name string) (*openstackv1beta1.OpenStackVersion, error) {
	unstructuredObj, err := c.client.Resource(openstackVersionGVR).
		Namespace(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenStackVersion: %w", err)
	}

	// Convert unstructured to OpenStackVersion
	data, err := unstructuredObj.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal unstructured object: %w", err)
	}

	var osVersion openstackv1beta1.OpenStackVersion
	if err := json.Unmarshal(data, &osVersion); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to OpenStackVersion: %w", err)
	}

	return &osVersion, nil
}

// ListOpenStackVersions lists all OpenStackVersion CRs in the specified namespace
func (c *K8sClient) ListOpenStackVersions(ctx context.Context, namespace string) ([]openstackv1beta1.OpenStackVersion, error) {
	unstructuredList, err := c.client.Resource(openstackVersionGVR).
		Namespace(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list OpenStackVersions: %w", err)
	}

	// Convert unstructured list to OpenStackVersionList
	data, err := unstructuredList.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal unstructured list: %w", err)
	}

	var osVersionList struct {
		Items []openstackv1beta1.OpenStackVersion `json:"items"`
	}
	if err := json.Unmarshal(data, &osVersionList); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to OpenStackVersionList: %w", err)
	}

	return osVersionList.Items, nil
}

// PatchOpenStackVersionTargetVersion patches the targetVersion field of an OpenStackVersion CR
func (c *K8sClient) PatchOpenStackVersionTargetVersion(ctx context.Context, namespace, name, targetVersion string) (*openstackv1beta1.OpenStackVersion, error) {
	// Create JSON patch for the targetVersion field
	patchData := []byte(fmt.Sprintf(`{"spec":{"targetVersion":"%s"}}`, targetVersion))

	// Apply the patch
	unstructuredObj, err := c.client.Resource(openstackVersionGVR).
		Namespace(namespace).
		Patch(ctx, name, "application/merge-patch+json", patchData, metav1.PatchOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to patch OpenStackVersion: %w", err)
	}

	// Convert unstructured to OpenStackVersion
	data, err := unstructuredObj.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal unstructured object: %w", err)
	}

	var osVersion openstackv1beta1.OpenStackVersion
	if err := json.Unmarshal(data, &osVersion); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to OpenStackVersion: %w", err)
	}

	return &osVersion, nil
}
