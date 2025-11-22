package handlers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dprince/openstack-k8s-mcp/internal/client"
	"github.com/mark3labs/mcp-go/mcp"
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

		// Convert response to JSON
		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// SetOpenStackVersionTargetVersionHandler handles the set_openstack_version_targetversion tool call
func SetOpenStackVersionTargetVersionHandler(k8sClient *client.K8sClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

		targetVersion, ok := request.Params.Arguments["targetVersion"].(string)
		if !ok || targetVersion == "" {
			return mcp.NewToolResultError("targetVersion parameter is required"), nil
		}

		// Patch the OpenStackVersion CR
		osVersion, err := k8sClient.PatchOpenStackVersionTargetVersion(ctx, namespace, name, targetVersion)
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
