package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dprince/openstack-k8s-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	openstackv1beta1 "github.com/openstack-k8s-operators/openstack-operator/apis/core/v1beta1"
)

const (
	DefaultNamespace = "openstack"
)

// GetOpenStackVersionHandler handles the get_openstack_version tool call
func GetOpenStackVersionHandler(k8sClient *client.K8sClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		namespace, ok := request.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			namespace = DefaultNamespace
		}

		name, ok := request.Params.Arguments["name"].(string)

		var osVersion *openstackv1beta1.OpenStackVersion
		var err error

		if ok && name != "" {
			// Query the specific OpenStackVersion CR by name
			osVersion, err = k8sClient.GetOpenStackVersion(ctx, namespace, name)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to get OpenStackVersion: %v", err)), nil
			}
		} else {
			// List all OpenStackVersion CRs and return the first one
			versions, err := k8sClient.ListOpenStackVersions(ctx, namespace)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to list OpenStackVersions: %v", err)), nil
			}

			if len(versions) == 0 {
				return mcp.NewToolResultError(fmt.Sprintf("No OpenStackVersion CR found in namespace %s", namespace)), nil
			}

			osVersion = &versions[0]
		}

		// Build response with full status information
		spec := map[string]interface{}{
			"targetVersion": osVersion.Spec.TargetVersion,
		}

		// Add customContainerImages if present
		if osVersion.Spec.CustomContainerImages.ContainerTemplate != (openstackv1beta1.ContainerTemplate{}) ||
			len(osVersion.Spec.CustomContainerImages.CinderVolumeImages) > 0 ||
			len(osVersion.Spec.CustomContainerImages.ManilaShareImages) > 0 {
			spec["customContainerImages"] = osVersion.Spec.CustomContainerImages
		}

		response := map[string]interface{}{
			"name":      osVersion.Name,
			"namespace": osVersion.Namespace,
			"spec":      spec,
			"status": map[string]interface{}{
				"availableVersion":  osVersion.Status.AvailableVersion,
				"deployedVersion":   osVersion.Status.DeployedVersion,
				"containerImages":   osVersion.Status.ContainerImages,
				"observedGeneration": osVersion.Status.ObservedGeneration,
			},
		}

		// Add conditions if present
		if len(osVersion.Status.Conditions) > 0 {
			conditions := make([]map[string]interface{}, len(osVersion.Status.Conditions))
			for i, cond := range osVersion.Status.Conditions {
				conditions[i] = map[string]interface{}{
					"type":               cond.Type,
					"status":             cond.Status,
					"lastTransitionTime": cond.LastTransitionTime.Time,
					"reason":             cond.Reason,
					"message":            cond.Message,
				}
				if cond.Severity != "" {
					conditions[i]["severity"] = cond.Severity
				}
			}
			response["status"].(map[string]interface{})["conditions"] = conditions
		}

		// Add containerImageVersionDefaults if present
		if osVersion.Status.ContainerImageVersionDefaults != nil {
			response["status"].(map[string]interface{})["containerImageVersionDefaults"] = osVersion.Status.ContainerImageVersionDefaults
		}

		// Add serviceDefaults if present
		response["status"].(map[string]interface{})["serviceDefaults"] = osVersion.Status.ServiceDefaults

		// Add availableServiceDefaults if present
		if osVersion.Status.AvailableServiceDefaults != nil {
			response["status"].(map[string]interface{})["availableServiceDefaults"] = osVersion.Status.AvailableServiceDefaults
		}

		// Add trackedCustomImages if present
		if osVersion.Status.TrackedCustomImages != nil {
			response["status"].(map[string]interface{})["trackedCustomImages"] = osVersion.Status.TrackedCustomImages
		}

		// Convert response to JSON
		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// UpdateOpenStackVersionHandler handles the update_openstack_version tool call
func UpdateOpenStackVersionHandler(k8sClient *client.K8sClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract parameters
		namespace, ok := request.Params.Arguments["namespace"].(string)
		if !ok || namespace == "" {
			namespace = DefaultNamespace
		}

		// Look up the first OpenStackVersion in the namespace
		versions, err := k8sClient.ListOpenStackVersions(ctx, namespace)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list OpenStackVersions: %v", err)), nil
		}

		if len(versions) == 0 {
			return mcp.NewToolResultError(fmt.Sprintf("No OpenStackVersion CR found in namespace %s", namespace)), nil
		}

		name := versions[0].Name

		targetVersion, ok := request.Params.Arguments["targetVersion"].(string)
		if !ok || targetVersion == "" {
			return mcp.NewToolResultError("targetVersion parameter is required"), nil
		}

		// Extract optional customContainerImages parameter
		var customContainerImages map[string]interface{}
		if customImages, ok := request.Params.Arguments["customContainerImages"].(map[string]interface{}); ok {
			customContainerImages = customImages
		}

		// Patch the OpenStackVersion CR
		osVersion, err := k8sClient.PatchOpenStackVersion(ctx, namespace, name, targetVersion, customContainerImages)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to patch OpenStackVersion: %v", err)), nil
		}

		// Build response with relevant fields
		response := map[string]interface{}{
			"name":      osVersion.Name,
			"namespace": osVersion.Namespace,
			"spec": map[string]interface{}{
				"targetVersion": osVersion.Spec.TargetVersion,
			},
			"status": map[string]interface{}{
				"availableVersion": osVersion.Status.AvailableVersion,
				"deployedVersion":  osVersion.Status.DeployedVersion,
			},
		}

		// Convert response to JSON
		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// WaitOpenStackVersionHandler handles the wait_openstack_version tool call
func WaitOpenStackVersionHandler(k8sClient *client.K8sClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		conditionType, ok := request.Params.Arguments["condition"].(string)
		if !ok || conditionType == "" {
			return mcp.NewToolResultError("condition parameter is required"), nil
		}

		// Optional timeout parameter (default 300 seconds)
		timeout := 300
		if timeoutVal, ok := request.Params.Arguments["timeout"].(float64); ok {
			timeout = int(timeoutVal)
		}

		// Optional pollInterval parameter (default 5 seconds)
		pollInterval := 5
		if pollIntervalVal, ok := request.Params.Arguments["pollInterval"].(float64); ok {
			pollInterval = int(pollIntervalVal)
		}

		// Get the MCP server from context to send log notifications
		mcpServer := server.ServerFromContext(ctx)

		// Create a logging function that sends notifications to the client
		logFunc := func(message string) {
			if mcpServer != nil {
				// Send a logging notification that will appear in the MCP client console
				_ = mcpServer.SendNotificationToClient("notifications/message", map[string]interface{}{
					"level":   "info",
					"message": message,
				})
			}
		}

		// Wait for the condition
		status, err := k8sClient.WaitForCondition(ctx, namespace, name, conditionType, timeout, pollInterval, logFunc)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to wait for condition: %v", err)), nil
		}

		// Build response
		response := map[string]interface{}{
			"name":      name,
			"namespace": namespace,
			"condition": conditionType,
			"met":       status.Met,
			"message":   status.Message,
			"reason":    status.Reason,
		}

		// Convert response to JSON
		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}
