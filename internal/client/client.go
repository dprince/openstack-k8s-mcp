package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

// ConditionStatus represents the result of checking a condition
type ConditionStatus struct {
	Met     bool
	Message string
	Reason  string
}

// WaitForCondition waits for a specific condition on an OpenStackVersion CR to become true
// logFunc is called periodically to provide status updates
func (c *K8sClient) WaitForCondition(ctx context.Context, namespace, name, conditionType string, timeoutSeconds int, pollIntervalSeconds int, logFunc func(string)) (*ConditionStatus, error) {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 300 // Default 5 minutes
	}

	if pollIntervalSeconds <= 0 {
		pollIntervalSeconds = 5 // Default 5 seconds
	}

	pollInterval := pollIntervalSeconds
	maxAttempts := timeoutSeconds / pollInterval

	logFunc(fmt.Sprintf("Waiting for condition '%s' on OpenStackVersion '%s/%s' (timeout: %ds)", conditionType, namespace, name, timeoutSeconds))

	for attempt := 0; attempt < maxAttempts; attempt++ {
		osVersion, err := c.GetOpenStackVersion(ctx, namespace, name)
		if err != nil {
			return nil, fmt.Errorf("failed to get OpenStackVersion: %w", err)
		}

		// Check if the condition exists and is true
		for _, cond := range osVersion.Status.Conditions {
			if string(cond.Type) == conditionType {
				if string(cond.Status) == "True" {
					logFunc(fmt.Sprintf("✓ Condition '%s' is True - Ready!", conditionType))
					return &ConditionStatus{
						Met:     true,
						Message: string(cond.Message),
						Reason:  string(cond.Reason),
					}, nil
				}
				// Condition exists but is not True - log current status
				logFunc(fmt.Sprintf("Polling... Condition '%s' status: %s (reason: %s)", conditionType, cond.Status, cond.Reason))
				break
			}
		}

		// Wait before next poll
		if attempt < maxAttempts-1 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(time.Duration(pollInterval) * time.Second):
				// Continue to next iteration
			}
		}
	}

	return &ConditionStatus{
		Met:     false,
		Message: fmt.Sprintf("Timeout waiting for condition '%s'", conditionType),
		Reason:  "Timeout",
	}, nil
}

// VerificationResult represents the result of verifying conditions
type VerificationResult struct {
	AllReady          bool
	TotalConditions   int
	ReadyConditions   []string
	NotReadyConditions []map[string]string
}

// VerifyControlPlaneConditions checks if all conditions on an OpenStackControlPlane CR are ready
func (c *K8sClient) VerifyControlPlaneConditions(ctx context.Context, namespace, name string) (*VerificationResult, error) {
	controlPlane, err := c.GetOpenStackControlPlane(ctx, namespace, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenStackControlPlane: %w", err)
	}

	result := &VerificationResult{
		AllReady:           true,
		ReadyConditions:    []string{},
		NotReadyConditions: []map[string]string{},
	}

	// Extract status and conditions
	status, ok := controlPlane["status"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("no status found on OpenStackControlPlane")
	}

	conditions, ok := status["conditions"].([]interface{})
	if !ok || len(conditions) == 0 {
		return nil, fmt.Errorf("no conditions found on OpenStackControlPlane")
	}

	result.TotalConditions = len(conditions)

	// Check all conditions
	for _, condInterface := range conditions {
		cond, ok := condInterface.(map[string]interface{})
		if !ok {
			continue
		}

		condType, _ := cond["type"].(string)
		condStatus, _ := cond["status"].(string)
		condReason, _ := cond["reason"].(string)
		condMessage, _ := cond["message"].(string)

		if condStatus == "True" {
			result.ReadyConditions = append(result.ReadyConditions, condType)
		} else {
			result.AllReady = false
			result.NotReadyConditions = append(result.NotReadyConditions, map[string]string{
				"type":    condType,
				"status":  condStatus,
				"reason":  condReason,
				"message": condMessage,
			})
		}
	}

	return result, nil
}

// NodeSetVerificationResult represents the verification result for a single NodeSet
type NodeSetVerificationResult struct {
	Name              string
	AllReady          bool
	TotalConditions   int
	ReadyConditions   []string
	NotReadyConditions []map[string]string
}

// AllNodeSetsVerificationResult represents the verification result for all NodeSets
type AllNodeSetsVerificationResult struct {
	AllReady  bool
	TotalNodeSets int
	ReadyNodeSets []string
	NotReadyNodeSets []NodeSetVerificationResult
}

// VerifyDataplaneNodeSetsConditions checks if all conditions on all OpenStackDataplaneNodeSet CRs are ready
func (c *K8sClient) VerifyDataplaneNodeSetsConditions(ctx context.Context, namespace string) (*AllNodeSetsVerificationResult, error) {
	nodeSets, err := c.ListDataplaneNodeSets(ctx, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to list OpenStackDataplaneNodeSets: %w", err)
	}

	if len(nodeSets) == 0 {
		return nil, fmt.Errorf("no OpenStackDataplaneNodeSets found in namespace '%s'", namespace)
	}

	result := &AllNodeSetsVerificationResult{
		AllReady:         true,
		TotalNodeSets:    len(nodeSets),
		ReadyNodeSets:    []string{},
		NotReadyNodeSets: []NodeSetVerificationResult{},
	}

	// Check each NodeSet
	for _, nodeSet := range nodeSets {
		metadata := nodeSet["metadata"].(map[string]interface{})
		name := metadata["name"].(string)

		nodeSetResult := NodeSetVerificationResult{
			Name:               name,
			AllReady:           true,
			ReadyConditions:    []string{},
			NotReadyConditions: []map[string]string{},
		}

		// Extract status and conditions
		status, ok := nodeSet["status"].(map[string]interface{})
		if !ok {
			result.AllReady = false
			nodeSetResult.AllReady = false
			nodeSetResult.NotReadyConditions = append(nodeSetResult.NotReadyConditions, map[string]string{
				"type":    "Status",
				"status":  "Unknown",
				"reason":  "NoStatus",
				"message": "No status found on NodeSet",
			})
			result.NotReadyNodeSets = append(result.NotReadyNodeSets, nodeSetResult)
			continue
		}

		conditions, ok := status["conditions"].([]interface{})
		if !ok || len(conditions) == 0 {
			result.AllReady = false
			nodeSetResult.AllReady = false
			nodeSetResult.NotReadyConditions = append(nodeSetResult.NotReadyConditions, map[string]string{
				"type":    "Conditions",
				"status":  "Unknown",
				"reason":  "NoConditions",
				"message": "No conditions found on NodeSet",
			})
			result.NotReadyNodeSets = append(result.NotReadyNodeSets, nodeSetResult)
			continue
		}

		nodeSetResult.TotalConditions = len(conditions)

		// Check all conditions for this NodeSet
		for _, condInterface := range conditions {
			cond, ok := condInterface.(map[string]interface{})
			if !ok {
				continue
			}

			condType, _ := cond["type"].(string)
			condStatus, _ := cond["status"].(string)
			condReason, _ := cond["reason"].(string)
			condMessage, _ := cond["message"].(string)

			if condStatus == "True" {
				nodeSetResult.ReadyConditions = append(nodeSetResult.ReadyConditions, condType)
			} else {
				nodeSetResult.AllReady = false
				result.AllReady = false
				nodeSetResult.NotReadyConditions = append(nodeSetResult.NotReadyConditions, map[string]string{
					"type":    condType,
					"status":  condStatus,
					"reason":  condReason,
					"message": condMessage,
				})
			}
		}

		// Add to appropriate list
		if nodeSetResult.AllReady {
			result.ReadyNodeSets = append(result.ReadyNodeSets, name)
		} else {
			result.NotReadyNodeSets = append(result.NotReadyNodeSets, nodeSetResult)
		}
	}

	return result, nil
}
