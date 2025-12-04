package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dprince/openstack-k8s-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// CreateDataplaneDeploymentHandler handles the create_dataplane_deployment tool call
func CreateDataplaneDeploymentHandler(k8sClient *client.K8sClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		namespace, ok := request.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			namespace = DefaultNamespace
		}

		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return newStructuredError(
				ErrorCodeInvalidParameter,
				"name parameter is required and must be a non-empty string",
				"ParameterValidationError",
			), nil
		}

		// Replace dots with dashes in the name
		name = strings.ReplaceAll(name, ".", "-")

		// Build the spec map - start with an empty map
		spec := make(map[string]interface{})

		// Check if a spec parameter was provided (new flexible approach)
		if specParam, ok := request.Params.Arguments["spec"].(map[string]interface{}); ok {
			// Use the provided spec directly
			spec = specParam
		} else {
			// Legacy approach: extract nodeSets and servicesOverride individually
			// Extract nodeSets parameter (optional - if not provided, auto-discover all nodeSets)
			nodeSetsRaw, ok := request.Params.Arguments["nodeSets"]

			var nodeSets []string
			if !ok {
				// No nodeSets provided - auto-discover all nodeSets in the namespace
				allNodeSets, err := k8sClient.ListDataplaneNodeSets(ctx, namespace)
				if err != nil {
					return newStructuredError(
						ErrorCodeK8sAPIError,
						fmt.Sprintf("Failed to list OpenStackDataplaneNodeSets in namespace '%s': %v", namespace, err),
						"KubernetesAPIError",
					), nil
				}

				if len(allNodeSets) == 0 {
					return newStructuredError(
						ErrorCodeInvalidParameter,
						fmt.Sprintf("No OpenStackDataplaneNodeSets found in namespace '%s'. Please create nodesets first or provide explicit nodeSets parameter.", namespace),
						"ParameterValidationError",
					), nil
				}

				// Extract names from all nodeSets
				nodeSets = make([]string, len(allNodeSets))
				for i, ns := range allNodeSets {
					metadata := ns["metadata"].(map[string]interface{})
					nodeSets[i] = metadata["name"].(string)
				}
			} else {
				// Convert provided nodeSets to []string
				nodeSetsArray, ok := nodeSetsRaw.([]interface{})
				if !ok {
					return newStructuredError(
						ErrorCodeInvalidParameter,
						"nodeSets must be an array",
						"ParameterValidationError",
					), nil
				}

				if len(nodeSetsArray) == 0 {
					return newStructuredError(
						ErrorCodeInvalidParameter,
						"nodeSets must contain at least one nodeSet",
						"ParameterValidationError",
					), nil
				}

				nodeSets = make([]string, len(nodeSetsArray))
				for i, ns := range nodeSetsArray {
					nodeSetStr, ok := ns.(string)
					if !ok {
						return newStructuredError(
							ErrorCodeInvalidParameter,
							fmt.Sprintf("nodeSets element at index %d must be a string", i),
							"ParameterValidationError",
						), nil
					}
					nodeSets[i] = nodeSetStr
				}
			}
			spec["nodeSets"] = nodeSets

			// Extract optional servicesOverride parameter
			if servicesOverrideRaw, ok := request.Params.Arguments["servicesOverride"]; ok {
				servicesOverrideArray, ok := servicesOverrideRaw.([]interface{})
				if !ok {
					return newStructuredError(
						ErrorCodeInvalidParameter,
						"servicesOverride must be an array of strings",
						"ParameterValidationError",
					), nil
				}

				servicesOverride := make([]string, len(servicesOverrideArray))
				for i, svc := range servicesOverrideArray {
					svcStr, ok := svc.(string)
					if !ok {
						return newStructuredError(
							ErrorCodeInvalidParameter,
							fmt.Sprintf("servicesOverride element at index %d must be a string", i),
							"ParameterValidationError",
						), nil
					}
					servicesOverride[i] = svcStr
				}
				spec["servicesOverride"] = servicesOverride
			}
		}

		// Validate that nodeSets is present in the spec
		if _, ok := spec["nodeSets"]; !ok {
			return newStructuredError(
				ErrorCodeInvalidParameter,
				"spec must contain 'nodeSets' field",
				"ParameterValidationError",
			), nil
		}

		// Set default deploymentRequeueTime to 1 if not provided
		if _, ok := spec["deploymentRequeueTime"]; !ok {
			spec["deploymentRequeueTime"] = 1
		}

		// Create the OpenStackDataplaneDeployment CR
		err := k8sClient.CreateDataplaneDeployment(ctx, namespace, name, spec)
		if err != nil {
			return newStructuredError(
				ErrorCodeK8sAPIError,
				fmt.Sprintf("Failed to create OpenStackDataplaneDeployment '%s' in namespace '%s': %v", name, namespace, err),
				"KubernetesAPIError",
			), nil
		}

		// Build success response
		specJSON, _ := json.MarshalIndent(spec, "", "  ")
		successMessage := fmt.Sprintf("Successfully created OpenStackDataplaneDeployment '%s' in namespace '%s' with spec:\n%s", name, namespace, string(specJSON))

		return mcp.NewToolResultText(successMessage), nil
	}
}

// GetDataplaneDeploymentHandler handles the get_dataplane_deployment tool call
func GetDataplaneDeploymentHandler(k8sClient *client.K8sClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		namespace, ok := request.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			namespace = DefaultNamespace
		}

		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return newStructuredError(
				ErrorCodeInvalidParameter,
				"name parameter is required and must be a non-empty string",
				"ParameterValidationError",
			), nil
		}

		// Replace dots with dashes in the name
		name = strings.ReplaceAll(name, ".", "-")

		// Get the OpenStackDataplaneDeployment CR
		deployment, err := k8sClient.GetDataplaneDeployment(ctx, namespace, name)
		if err != nil {
			return newStructuredError(
				ErrorCodeK8sAPIError,
				fmt.Sprintf("Failed to get OpenStackDataplaneDeployment '%s' in namespace '%s': %v", name, namespace, err),
				"KubernetesAPIError",
			), nil
		}

		// Extract relevant fields
		response := map[string]interface{}{
			"name":      deployment["metadata"].(map[string]interface{})["name"],
			"namespace": deployment["metadata"].(map[string]interface{})["namespace"],
		}

		// Add spec fields
		if spec, ok := deployment["spec"].(map[string]interface{}); ok {
			response["spec"] = spec
		}

		// Add status fields if present
		if status, ok := deployment["status"].(map[string]interface{}); ok {
			response["status"] = status
		}

		// Convert response to JSON
		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return newStructuredError(
				ErrorCodeMarshalError,
				fmt.Sprintf("Failed to marshal response: %v", err),
				"MarshalError",
			), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// CreateDataplaneDeploymentOVNHandler handles the create_dataplane_deployment_ovn tool call
func CreateDataplaneDeploymentOVNHandler(k8sClient *client.K8sClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		namespace, ok := request.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			namespace = DefaultNamespace
		}

		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return newStructuredError(
				ErrorCodeInvalidParameter,
				"name parameter is required and must be a non-empty string",
				"ParameterValidationError",
			), nil
		}

		// Replace dots with dashes in the name
		name = strings.ReplaceAll(name, ".", "-")

		// Build the spec map
		spec := make(map[string]interface{})

		// Auto-discover all nodeSets in the namespace
		allNodeSets, err := k8sClient.ListDataplaneNodeSets(ctx, namespace)
		if err != nil {
			return newStructuredError(
				ErrorCodeK8sAPIError,
				fmt.Sprintf("Failed to list OpenStackDataplaneNodeSets in namespace '%s': %v", namespace, err),
				"KubernetesAPIError",
			), nil
		}

		if len(allNodeSets) == 0 {
			return newStructuredError(
				ErrorCodeInvalidParameter,
				fmt.Sprintf("No OpenStackDataplaneNodeSets found in namespace '%s'. Please create nodesets first.", namespace),
				"ParameterValidationError",
			), nil
		}

		// Extract names from all nodeSets
		nodeSets := make([]string, len(allNodeSets))
		for i, ns := range allNodeSets {
			metadata := ns["metadata"].(map[string]interface{})
			nodeSets[i] = metadata["name"].(string)
		}
		spec["nodeSets"] = nodeSets

		// Set servicesOverride to ["ovn"]
		spec["servicesOverride"] = []string{"ovn"}

		// Set default deploymentRequeueTime to 1
		spec["deploymentRequeueTime"] = 1

		// Create the OpenStackDataplaneDeployment CR
		err = k8sClient.CreateDataplaneDeployment(ctx, namespace, name, spec)
		if err != nil {
			return newStructuredError(
				ErrorCodeK8sAPIError,
				fmt.Sprintf("Failed to create OpenStackDataplaneDeployment '%s' in namespace '%s': %v", name, namespace, err),
				"KubernetesAPIError",
			), nil
		}

		// Build success response
		specJSON, _ := json.MarshalIndent(spec, "", "  ")
		successMessage := fmt.Sprintf("Successfully created OpenStackDataplaneDeployment '%s' in namespace '%s' with servicesOverride=[ovn] and spec:\n%s", name, namespace, string(specJSON))

		return mcp.NewToolResultText(successMessage), nil
	}
}

// CreateDataplaneDeploymentUpdateHandler handles the create_dataplane_deployment_update tool call
func CreateDataplaneDeploymentUpdateHandler(k8sClient *client.K8sClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		namespace, ok := request.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			namespace = DefaultNamespace
		}

		name, ok := request.Params.Arguments["name"].(string)
		if !ok || name == "" {
			return newStructuredError(
				ErrorCodeInvalidParameter,
				"name parameter is required and must be a non-empty string",
				"ParameterValidationError",
			), nil
		}

		// Replace dots with dashes in the name
		name = strings.ReplaceAll(name, ".", "-")

		// Build the spec map
		spec := make(map[string]interface{})

		// Auto-discover all nodeSets in the namespace
		allNodeSets, err := k8sClient.ListDataplaneNodeSets(ctx, namespace)
		if err != nil {
			return newStructuredError(
				ErrorCodeK8sAPIError,
				fmt.Sprintf("Failed to list OpenStackDataplaneNodeSets in namespace '%s': %v", namespace, err),
				"KubernetesAPIError",
			), nil
		}

		if len(allNodeSets) == 0 {
			return newStructuredError(
				ErrorCodeInvalidParameter,
				fmt.Sprintf("No OpenStackDataplaneNodeSets found in namespace '%s'. Please create nodesets first.", namespace),
				"ParameterValidationError",
			), nil
		}

		// Extract names from all nodeSets
		nodeSets := make([]string, len(allNodeSets))
		for i, ns := range allNodeSets {
			metadata := ns["metadata"].(map[string]interface{})
			nodeSets[i] = metadata["name"].(string)
		}
		spec["nodeSets"] = nodeSets

		// Set servicesOverride to ["update"]
		spec["servicesOverride"] = []string{"update"}

		// Set default deploymentRequeueTime to 1
		spec["deploymentRequeueTime"] = 1

		// Create the OpenStackDataplaneDeployment CR
		err = k8sClient.CreateDataplaneDeployment(ctx, namespace, name, spec)
		if err != nil {
			return newStructuredError(
				ErrorCodeK8sAPIError,
				fmt.Sprintf("Failed to create OpenStackDataplaneDeployment '%s' in namespace '%s': %v", name, namespace, err),
				"KubernetesAPIError",
			), nil
		}

		// Build success response
		specJSON, _ := json.MarshalIndent(spec, "", "  ")
		successMessage := fmt.Sprintf("Successfully created OpenStackDataplaneDeployment '%s' in namespace '%s' with servicesOverride=[update] and spec:\n%s", name, namespace, string(specJSON))

		return mcp.NewToolResultText(successMessage), nil
	}
}

// ListDataplaneDeploymentsHandler handles the list_dataplane_deployments tool call
func ListDataplaneDeploymentsHandler(k8sClient *client.K8sClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		namespace, ok := request.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			namespace = DefaultNamespace
		}

		// List all OpenStackDataplaneDeployment CRs
		deployments, err := k8sClient.ListDataplaneDeployments(ctx, namespace)
		if err != nil {
			return newStructuredError(
				ErrorCodeK8sAPIError,
				fmt.Sprintf("Failed to list OpenStackDataplaneDeployments in namespace '%s': %v", namespace, err),
				"KubernetesAPIError",
			), nil
		}

		if len(deployments) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No OpenStackDataplaneDeployments found in namespace '%s'", namespace)), nil
		}

		// Build response with relevant fields from each deployment
		response := make([]map[string]interface{}, len(deployments))
		for i, deployment := range deployments {
			metadata := deployment["metadata"].(map[string]interface{})
			item := map[string]interface{}{
				"name":      metadata["name"],
				"namespace": metadata["namespace"],
			}

			// Add spec fields if present
			if spec, ok := deployment["spec"].(map[string]interface{}); ok {
				item["spec"] = spec
			}

			// Add status fields if present
			if status, ok := deployment["status"].(map[string]interface{}); ok {
				item["status"] = status
			}

			response[i] = item
		}

		// Convert response to JSON
		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return newStructuredError(
				ErrorCodeMarshalError,
				fmt.Sprintf("Failed to marshal response: %v", err),
				"MarshalError",
			), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
