package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dprince/openstack-k8s-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
)

// GetOpenStackControlPlaneHandler handles the get_openstack_controlplane tool call
func GetOpenStackControlPlaneHandler(k8sClient *client.K8sClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		namespace, ok := request.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			namespace = DefaultNamespace
		}

		name, ok := request.Params.Arguments["name"].(string)

		var controlPlane map[string]interface{}
		var err error

		if ok && name != "" {
			// Query the specific OpenStackControlPlane CR by name
			controlPlane, err = k8sClient.GetOpenStackControlPlane(ctx, namespace, name)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to get OpenStackControlPlane '%s' in namespace '%s': %v", name, namespace, err)), nil
			}
		} else {
			// Auto-discover: List all OpenStackControlPlane CRs and return the first one
			controlPlanes, err := k8sClient.ListOpenStackControlPlanes(ctx, namespace)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to list OpenStackControlPlanes in namespace '%s': %v", namespace, err)), nil
			}

			if len(controlPlanes) == 0 {
				return mcp.NewToolResultError(fmt.Sprintf("No OpenStackControlPlane CR found in namespace '%s'", namespace)), nil
			}

			controlPlane = controlPlanes[0]
		}

		// Extract relevant fields
		metadata := controlPlane["metadata"].(map[string]interface{})
		response := map[string]interface{}{
			"name":      metadata["name"],
			"namespace": metadata["namespace"],
		}

		// Add spec fields
		if spec, ok := controlPlane["spec"].(map[string]interface{}); ok {
			response["spec"] = spec
		}

		// Add status fields if present
		if status, ok := controlPlane["status"].(map[string]interface{}); ok {
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

// VerifyOpenStackControlPlaneHandler handles the verify_openstack_controlplane tool call
func VerifyOpenStackControlPlaneHandler(k8sClient *client.K8sClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		namespace, ok := request.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			namespace = DefaultNamespace
		}

		name, ok := request.Params.Arguments["name"].(string)

		// If name is not provided, auto-discover the first OpenStackControlPlane in the namespace
		if !ok || name == "" {
			controlPlanes, err := k8sClient.ListOpenStackControlPlanes(ctx, namespace)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to list OpenStackControlPlanes in namespace '%s': %v", namespace, err)), nil
			}

			if len(controlPlanes) == 0 {
				return mcp.NewToolResultError(fmt.Sprintf("No OpenStackControlPlane CR found in namespace '%s'", namespace)), nil
			}

			metadata := controlPlanes[0]["metadata"].(map[string]interface{})
			name = metadata["name"].(string)
		}

		// Verify all conditions are ready
		result, err := k8sClient.VerifyControlPlaneConditions(ctx, namespace, name)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to verify OpenStackControlPlane '%s' in namespace '%s': %v", name, namespace, err)), nil
		}

		// Build response
		response := map[string]interface{}{
			"name":              name,
			"namespace":         namespace,
			"allReady":          result.AllReady,
			"totalConditions":   result.TotalConditions,
			"readyConditions":   result.ReadyConditions,
			"notReadyConditions": result.NotReadyConditions,
		}

		// Convert response to JSON
		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
