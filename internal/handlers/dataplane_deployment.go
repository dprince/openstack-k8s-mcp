package handlers

import (
	"context"
	"encoding/json"
	"fmt"

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
			return mcp.NewToolResultError("name parameter is required"), nil
		}

		// Build the spec map - start with an empty map
		spec := make(map[string]interface{})

		// Check if a spec parameter was provided (new flexible approach)
		if specParam, ok := request.Params.Arguments["spec"].(map[string]interface{}); ok {
			// Use the provided spec directly
			spec = specParam
		} else {
			// Legacy approach: extract nodeSets and servicesOverride individually
			// Extract nodeSets parameter (required when not using spec, must be array)
			nodeSetsRaw, ok := request.Params.Arguments["nodeSets"]
			if !ok {
				return mcp.NewToolResultError("either 'spec' parameter or 'nodeSets' parameter is required"), nil
			}

			// Convert nodeSets to []string
			nodeSetsArray, ok := nodeSetsRaw.([]interface{})
			if !ok {
				return mcp.NewToolResultError("nodeSets must be an array"), nil
			}

			if len(nodeSetsArray) == 0 {
				return mcp.NewToolResultError("nodeSets must contain at least one nodeSet"), nil
			}

			nodeSets := make([]string, len(nodeSetsArray))
			for i, ns := range nodeSetsArray {
				nodeSetStr, ok := ns.(string)
				if !ok {
					return mcp.NewToolResultError(fmt.Sprintf("nodeSets element at index %d must be a string", i)), nil
				}
				nodeSets[i] = nodeSetStr
			}
			spec["nodeSets"] = nodeSets

			// Extract optional servicesOverride parameter
			if servicesOverrideRaw, ok := request.Params.Arguments["servicesOverride"]; ok {
				servicesOverrideArray, ok := servicesOverrideRaw.([]interface{})
				if !ok {
					return mcp.NewToolResultError("servicesOverride must be an array of strings"), nil
				}

				servicesOverride := make([]string, len(servicesOverrideArray))
				for i, svc := range servicesOverrideArray {
					svcStr, ok := svc.(string)
					if !ok {
						return mcp.NewToolResultError(fmt.Sprintf("servicesOverride element at index %d must be a string", i)), nil
					}
					servicesOverride[i] = svcStr
				}
				spec["servicesOverride"] = servicesOverride
			}
		}

		// Validate that nodeSets is present in the spec
		if _, ok := spec["nodeSets"]; !ok {
			return mcp.NewToolResultError("spec must contain 'nodeSets' field"), nil
		}

		// Create the OpenStackDataplaneDeployment CR
		err := k8sClient.CreateDataplaneDeployment(ctx, namespace, name, spec)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to create OpenStackDataplaneDeployment: %v", err)), nil
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
			return mcp.NewToolResultError("name parameter is required"), nil
		}

		// Get the OpenStackDataplaneDeployment CR
		deployment, err := k8sClient.GetDataplaneDeployment(ctx, namespace, name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to get OpenStackDataplaneDeployment: %v", err)), nil
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
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
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
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list OpenStackDataplaneDeployments: %v", err)), nil
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
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
