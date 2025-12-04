package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dprince/openstack-k8s-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListDataplaneNodeSetsHandler handles the list_dataplane_nodesets tool call
func ListDataplaneNodeSetsHandler(k8sClient *client.K8sClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		namespace, ok := request.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			namespace = DefaultNamespace
		}

		// List all OpenStackDataplaneNodeSet CRs
		nodeSets, err := k8sClient.ListDataplaneNodeSets(ctx, namespace)
		if err != nil {
			return newStructuredError(
				ErrorCodeK8sAPIError,
				fmt.Sprintf("Failed to list OpenStackDataplaneNodeSets in namespace '%s': %v", namespace, err),
				"KubernetesAPIError",
			), nil
		}

		if len(nodeSets) == 0 {
			return mcp.NewToolResultText(fmt.Sprintf("No OpenStackDataplaneNodeSets found in namespace '%s'", namespace)), nil
		}

		// Build response with relevant fields from each nodeSet
		response := make([]map[string]interface{}, len(nodeSets))
		for i, nodeSet := range nodeSets {
			metadata := nodeSet["metadata"].(map[string]interface{})
			item := map[string]interface{}{
				"name":      metadata["name"],
				"namespace": metadata["namespace"],
			}

			// Add spec fields if present
			if spec, ok := nodeSet["spec"].(map[string]interface{}); ok {
				item["spec"] = spec
			}

			// Add status fields if present
			if status, ok := nodeSet["status"].(map[string]interface{}); ok {
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

// VerifyDataplaneNodeSetsHandler handles the verify_openstack_dataplanenodesets tool call
func VerifyDataplaneNodeSetsHandler(k8sClient *client.K8sClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		namespace, ok := request.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			namespace = DefaultNamespace
		}

		// Verify all NodeSets in the namespace
		result, err := k8sClient.VerifyDataplaneNodeSetsConditions(ctx, namespace)
		if err != nil {
			return newStructuredError(
				ErrorCodeK8sAPIError,
				fmt.Sprintf("Failed to verify OpenStackDataplaneNodeSets in namespace '%s': %v", namespace, err),
				"KubernetesAPIError",
			), nil
		}

		// Build response
		response := map[string]interface{}{
			"namespace":        namespace,
			"allReady":         result.AllReady,
			"totalNodeSets":    result.TotalNodeSets,
			"readyNodeSets":    result.ReadyNodeSets,
			"notReadyNodeSets": result.NotReadyNodeSets,
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
