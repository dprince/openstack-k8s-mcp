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

// Error codes for structured error responses
const (
	ErrorCodeNotFound         = "RESOURCE_NOT_FOUND"
	ErrorCodeInvalidParameter = "INVALID_PARAMETER"
	ErrorCodeK8sAPIError      = "K8S_API_ERROR"
	ErrorCodeMarshalError     = "MARSHAL_ERROR"
	ErrorCodeTimeout          = "TIMEOUT"
	ErrorCodeConditionNotMet  = "CONDITION_NOT_MET"
)

// StructuredError represents a structured error response for better LLM parsing
type StructuredError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

// newStructuredError creates a structured error response
func newStructuredError(code, message, errType string) *mcp.CallToolResult {
	errData := StructuredError{
		Code:    code,
		Message: message,
		Type:    errType,
	}
	jsonData, _ := json.Marshal(errData)
	return mcp.NewToolResultError(string(jsonData))
}

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
				return mcp.NewToolResultError(fmt.Sprintf("Failed to get OpenStackVersion '%s' in namespace '%s': %v", name, namespace, err)), nil
			}
		} else {
			// Auto-discover: List all OpenStackVersion CRs and return the first one
			versions, err := k8sClient.ListOpenStackVersions(ctx, namespace)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to list OpenStackVersions in namespace '%s': %v", namespace, err)), nil
			}

			if len(versions) == 0 {
				return mcp.NewToolResultError(fmt.Sprintf("No OpenStackVersion CR found in namespace '%s'", namespace)), nil
			}

			osVersion = &versions[0]
		}

		// Build flattened response
		response := map[string]interface{}{
			"name":             osVersion.Name,
			"namespace":        osVersion.Namespace,
			"targetVersion":    osVersion.Spec.TargetVersion,
			"availableVersion": osVersion.Status.AvailableVersion,
			"deployedVersion":  osVersion.Status.DeployedVersion,
		}

		// Add customContainerImages if present
		if osVersion.Spec.CustomContainerImages.ContainerTemplate != (openstackv1beta1.ContainerTemplate{}) ||
			len(osVersion.Spec.CustomContainerImages.CinderVolumeImages) > 0 ||
			len(osVersion.Spec.CustomContainerImages.ManilaShareImages) > 0 {
			response["customContainerImages"] = osVersion.Spec.CustomContainerImages
		}

		// Process conditions into ready and notReady arrays
		readyConditions := []string{}
		notReadyConditions := []string{}

		for _, cond := range osVersion.Status.Conditions {
			if cond.Status == "True" {
				readyConditions = append(readyConditions, string(cond.Type))
			} else {
				notReadyConditions = append(notReadyConditions, string(cond.Type))
			}
		}

		response["readyConditions"] = readyConditions
		response["notReadyConditions"] = notReadyConditions

		// Add optional detailed fields if present
		// if osVersion.Status.ContainerImages != nil {
		// 	response["containerImages"] = osVersion.Status.ContainerImages
		// }

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

		// Auto-discover the first OpenStackVersion in the namespace
		versions, err := k8sClient.ListOpenStackVersions(ctx, namespace)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to list OpenStackVersions in namespace '%s': %v", namespace, err)), nil
		}

		if len(versions) == 0 {
			return mcp.NewToolResultError(fmt.Sprintf("No OpenStackVersion CR found in namespace '%s'", namespace)), nil
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
			return mcp.NewToolResultError(fmt.Sprintf("Failed to patch OpenStackVersion '%s' in namespace '%s': %v", name, namespace, err)), nil
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

		// If name is not provided, auto-discover the first OpenStackVersion in the namespace
		if !ok || name == "" {
			versions, err := k8sClient.ListOpenStackVersions(ctx, namespace)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to list OpenStackVersions in namespace '%s': %v", namespace, err)), nil
			}

			if len(versions) == 0 {
				return mcp.NewToolResultError(fmt.Sprintf("No OpenStackVersion CR found in namespace '%s'", namespace)), nil
			}

			name = versions[0].Name
		}

		conditionType, ok := request.Params.Arguments["condition"].(string)
		if !ok || conditionType == "" {
			return mcp.NewToolResultError("condition parameter is required"), nil
		}

		// Optional timeout parameter (default 600 seconds)
		timeout := 600
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
			return mcp.NewToolResultError(fmt.Sprintf("Failed to wait for condition '%s' on '%s' in namespace '%s': %v", conditionType, name, namespace, err)), nil
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

// GetResumeStepHandler determines which upgrade step to resume from
func GetResumeStepHandler(k8sClient *client.K8sClient) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
				return mcp.NewToolResultError(fmt.Sprintf("Failed to get OpenStackVersion '%s' in namespace '%s': %v", name, namespace, err)), nil
			}
		} else {
			// Auto-discover: List all OpenStackVersion CRs and return the first one
			versions, err := k8sClient.ListOpenStackVersions(ctx, namespace)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("Failed to list OpenStackVersions in namespace '%s': %v", namespace, err)), nil
			}

			if len(versions) == 0 {
				return mcp.NewToolResultError(fmt.Sprintf("No OpenStackVersion CR found in namespace '%s'", namespace)), nil
			}

			osVersion = &versions[0]
		}

		// Extract version information
		targetVersion := osVersion.Spec.TargetVersion
		availableVersion := osVersion.Status.AvailableVersion
		deployedVersion := osVersion.Status.DeployedVersion

		// Build notReadyConditions array
		notReadyConditions := []string{}
		for _, cond := range osVersion.Status.Conditions {
			if cond.Status != "True" {
				notReadyConditions = append(notReadyConditions, string(cond.Type))
			}
		}

		// Determine the resume step based on the decision logic
		var resumeStep int
		var explanation string

		// Check if upgrade is in progress
		if availableVersion == nil || targetVersion != *availableVersion {
			// Not in progress - should start from pre-upgrade validation
			resumeStep = 2
			availableVersionStr := "nil"
			if availableVersion != nil {
				availableVersionStr = *availableVersion
			}
			explanation = fmt.Sprintf("Upgrade not in progress (targetVersion='%s' != availableVersion='%s'). Start from Step 2: Pre-Upgrade Validation.", targetVersion, availableVersionStr)
		} else if deployedVersion != nil && targetVersion == *deployedVersion && len(notReadyConditions) == 0 {
			// All conditions ready and target equals deployed - upgrade is complete
			resumeStep = 10
			explanation = fmt.Sprintf("Upgrade complete (targetVersion='%s' == deployedVersion='%s' and all conditions ready). Jump to Step 10: Update Complete.", targetVersion, *deployedVersion)
		} else {
			// Upgrade is in progress - check notReadyConditions
			upgradeInProgress := true

			// Map conditions to steps based on the resume decision table
			if contains(notReadyConditions, "MinorUpdateOVNControlplane") {
				resumeStep = 4
				explanation = "notReadyConditions contains 'MinorUpdateOVNControlplane'. Resume at Step 4: Monitor OVN Controlplane Deployment."
			} else if contains(notReadyConditions, "MinorUpdateOVNDataplane") {
				resumeStep = 5
				explanation = "notReadyConditions contains 'MinorUpdateOVNDataplane'. Resume at Step 5: Deploy OVN on Dataplane."
			} else if contains(notReadyConditions, "MinorUpdateControlplane") {
				resumeStep = 7
				explanation = "notReadyConditions contains 'MinorUpdateControlplane'. Resume at Step 7: Monitor Controlplane Update Completion."
			} else if contains(notReadyConditions, "MinorUpdateDataplane") {
				resumeStep = 8
				explanation = "notReadyConditions contains 'MinorUpdateDataplane'. Resume at Step 8: Deploy Update on Dataplane."
			} else if len(notReadyConditions) == 0 && deployedVersion != nil && targetVersion == *deployedVersion {
				resumeStep = 10
				explanation = "All conditions ready and targetVersion equals deployedVersion. Resume at Step 10: Update Complete."
			} else {
				// Fallback - continue with pre-upgrade validation
				resumeStep = 2
				explanation = fmt.Sprintf("Could not determine specific resume point from notReadyConditions=%v. Starting from Step 2: Pre-Upgrade Validation.", notReadyConditions)
				upgradeInProgress = false
			}

			if upgradeInProgress {
				availableVersionStr := "nil"
				if availableVersion != nil {
					availableVersionStr = *availableVersion
				}
				explanation = fmt.Sprintf("Upgrade in progress (targetVersion='%s' == availableVersion='%s'). %s", targetVersion, availableVersionStr, explanation)
			}
		}

		// Build response
		response := map[string]interface{}{
			"name":               osVersion.Name,
			"namespace":          osVersion.Namespace,
			"targetVersion":      targetVersion,
			"availableVersion":   availableVersion,
			"deployedVersion":    deployedVersion,
			"notReadyConditions": notReadyConditions,
			"resumeStep":         resumeStep,
			"explanation":        explanation,
		}

		// Convert response to JSON
		jsonData, err := json.MarshalIndent(response, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// contains checks if a string slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
