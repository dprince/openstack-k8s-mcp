package client

import (
	"context"
	"encoding/json"
	"fmt"

	openstackv1beta1 "github.com/openstack-k8s-operators/openstack-operator/apis/core/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
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

	openstackControlPlaneGVR = schema.GroupVersionResource{
		Group:    "core.openstack.org",
		Version:  "v1beta1",
		Resource: "openstackcontrolplanes",
	}

	openstackDataplaneDeploymentGVR = schema.GroupVersionResource{
		Group:    "dataplane.openstack.org",
		Version:  "v1beta1",
		Resource: "openstackdataplanedeployments",
	}

	openstackDataplaneNodeSetGVR = schema.GroupVersionResource{
		Group:    "dataplane.openstack.org",
		Version:  "v1beta1",
		Resource: "openstackdataplanenodesets",
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

// GetOpenStackControlPlane retrieves OpenStackControlPlane CR from the specified namespace
func (c *K8sClient) GetOpenStackControlPlane(ctx context.Context, namespace, name string) (map[string]interface{}, error) {
	unstructuredObj, err := c.client.Resource(openstackControlPlaneGVR).
		Namespace(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenStackControlPlane: %w", err)
	}

	return unstructuredObj.Object, nil
}

// ListOpenStackControlPlanes lists all OpenStackControlPlane CRs in the specified namespace
func (c *K8sClient) ListOpenStackControlPlanes(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	unstructuredList, err := c.client.Resource(openstackControlPlaneGVR).
		Namespace(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list OpenStackControlPlanes: %w", err)
	}

	controlPlanes := make([]map[string]interface{}, len(unstructuredList.Items))
	for i, item := range unstructuredList.Items {
		controlPlanes[i] = item.Object
	}

	return controlPlanes, nil
}

// PatchOpenStackVersion patches the targetVersion and optionally customContainerImages fields of an OpenStackVersion CR
func (c *K8sClient) PatchOpenStackVersion(ctx context.Context, namespace, name, targetVersion string, customContainerImages map[string]interface{}) (*openstackv1beta1.OpenStackVersion, error) {
	// Build the patch data structure
	spec := map[string]interface{}{
		"targetVersion": targetVersion,
	}

	// Add customContainerImages to spec if provided
	if customContainerImages != nil && len(customContainerImages) > 0 {
		spec["customContainerImages"] = customContainerImages
	}

	patch := map[string]interface{}{
		"spec": spec,
	}

	// Marshal the patch to JSON
	patchData, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal patch data: %w", err)
	}

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

// CreateDataplaneDeployment creates a new OpenStackDataplaneDeployment CR
func (c *K8sClient) CreateDataplaneDeployment(ctx context.Context, namespace, name string, spec map[string]interface{}) error {
	// Build the deployment object
	deployment := map[string]interface{}{
		"apiVersion": "dataplane.openstack.org/v1beta1",
		"kind":       "OpenStackDataplaneDeployment",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": spec,
	}

	// Marshal to JSON for creation
	deploymentJSON, err := json.Marshal(deployment)
	if err != nil {
		return fmt.Errorf("failed to marshal deployment: %w", err)
	}

	// Convert to unstructured
	var unstructuredDeployment map[string]interface{}
	if err := json.Unmarshal(deploymentJSON, &unstructuredDeployment); err != nil {
		return fmt.Errorf("failed to unmarshal to unstructured: %w", err)
	}

	// Create the deployment
	_, err = c.client.Resource(openstackDataplaneDeploymentGVR).
		Namespace(namespace).
		Create(ctx, &unstructured.Unstructured{Object: unstructuredDeployment}, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create OpenStackDataplaneDeployment: %w", err)
	}

	return nil
}

// GetDataplaneDeployment retrieves an OpenStackDataplaneDeployment CR from the specified namespace
func (c *K8sClient) GetDataplaneDeployment(ctx context.Context, namespace, name string) (map[string]interface{}, error) {
	unstructuredObj, err := c.client.Resource(openstackDataplaneDeploymentGVR).
		Namespace(namespace).
		Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenStackDataplaneDeployment: %w", err)
	}

	return unstructuredObj.Object, nil
}

// ListDataplaneDeployments lists all OpenStackDataplaneDeployment CRs in the specified namespace
func (c *K8sClient) ListDataplaneDeployments(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	unstructuredList, err := c.client.Resource(openstackDataplaneDeploymentGVR).
		Namespace(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list OpenStackDataplaneDeployments: %w", err)
	}

	deployments := make([]map[string]interface{}, len(unstructuredList.Items))
	for i, item := range unstructuredList.Items {
		deployments[i] = item.Object
	}

	return deployments, nil
}

// ListDataplaneNodeSets lists all OpenStackDataplaneNodeSet CRs in the specified namespace
func (c *K8sClient) ListDataplaneNodeSets(ctx context.Context, namespace string) ([]map[string]interface{}, error) {
	unstructuredList, err := c.client.Resource(openstackDataplaneNodeSetGVR).
		Namespace(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list OpenStackDataplaneNodeSets: %w", err)
	}

	nodeSets := make([]map[string]interface{}, len(unstructuredList.Items))
	for i, item := range unstructuredList.Items {
		nodeSets[i] = item.Object
	}

	return nodeSets, nil
}
