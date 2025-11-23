package main

import (
	"fmt"
	"log"
	"os"

	"github.com/dprince/openstack-k8s-mcp/internal/client"
	"github.com/dprince/openstack-k8s-mcp/internal/handlers"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Initialize Kubernetes client
	k8sClient, err := client.NewK8sClient()
	if err != nil {
		log.Fatalf("Failed to create Kubernetes client: %v", err)
	}

	// Create MCP server
	s := server.NewMCPServer(
		"openstack-k8s-mcp",
		"1.0.0",
	)

	// Register the get_openstack_version tool
	getOpenStackVersionTool := mcp.NewTool("get_openstack_version",
		mcp.WithDescription("Query OpenStackVersion CRD to get targetVersion, availableVersion, and conditions. If name is not provided, returns the first OpenStackVersion CR found in the namespace."),
		mcp.WithString("namespace",
			mcp.Required(),
			mcp.Description("Kubernetes namespace where the OpenStackVersion CR is located"),
		),
		mcp.WithString("name",
			mcp.Description("Name of the OpenStackVersion CR to query. If not provided, returns the first CR found in the namespace."),
		),
	)

	s.AddTool(getOpenStackVersionTool, handlers.GetOpenStackVersionHandler(k8sClient))

	// Register the update_openstack_version tool
	updateOpenStackVersionTool := mcp.NewTool("update_openstack_version",
		mcp.WithDescription("Patch the targetVersion and optionally customContainerImages fields of an OpenStackVersion CR. The customContainerImages parameter is optional and should be a map of service names to container image URLs."),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace where the OpenStackVersion CR is located. Defaults to 'openstack' if not provided."),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the OpenStackVersion CR to patch"),
		),
		mcp.WithString("targetVersion",
			mcp.Required(),
			mcp.Description("The target version to set for the OpenStackVersion CR"),
		),
	)

	s.AddTool(updateOpenStackVersionTool, handlers.UpdateOpenStackVersionHandler(k8sClient))

	// Register the wait_openstack_version tool
	waitOpenStackVersionTool := mcp.NewTool("wait_openstack_version",
		mcp.WithDescription("Wait for a specific condition on an OpenStackVersion CR to become True. This tool polls the CR status and provides periodic progress updates. Useful for waiting on conditions like 'MinorUpdateReady', 'Ready', etc."),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace where the OpenStackVersion CR is located. Defaults to 'openstack' if not provided."),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the OpenStackVersion CR to monitor"),
		),
		mcp.WithString("condition",
			mcp.Required(),
			mcp.Description("The condition type to wait for (e.g., 'MinorUpdateReady', 'Ready'). The tool waits until this condition's status becomes 'True'."),
		),
		mcp.WithNumber("timeout",
			mcp.Description("Timeout in seconds to wait for the condition. Defaults to 300 seconds (5 minutes) if not provided."),
		),
		mcp.WithNumber("pollInterval",
			mcp.Description("Interval in seconds between polling attempts. Defaults to 5 seconds if not provided."),
		),
	)

	s.AddTool(waitOpenStackVersionTool, handlers.WaitOpenStackVersionHandler(k8sClient))

	// Register the get_openstack_controlplane tool
	getOpenStackControlPlaneTool := mcp.NewTool("get_openstack_controlplane",
		mcp.WithDescription("Query OpenStackControlPlane CRD to get spec and status information. If name is not provided, returns the first OpenStackControlPlane CR found in the namespace."),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace where the OpenStackControlPlane CR is located. Defaults to 'openstack' if not provided."),
		),
		mcp.WithString("name",
			mcp.Description("Name of the OpenStackControlPlane CR to query. If not provided, returns the first CR found in the namespace."),
		),
	)

	s.AddTool(getOpenStackControlPlaneTool, handlers.GetOpenStackControlPlaneHandler(k8sClient))

	// Register the create_dataplane_deployment tool
	// Note: The spec parameter accepts any valid OpenStackDataplaneDeployment spec fields as a JSON object
	// Alternatively, nodeSets and servicesOverride can be passed as individual parameters for backward compatibility
	createDataplaneDeploymentTool := mcp.NewTool("create_dataplane_deployment",
		mcp.WithDescription("Create an OpenStackDataplaneDeployment CR with a flexible spec. You can either:\n1. Pass a 'spec' parameter containing all spec fields (nodeSets, servicesOverride, ansibleTags, ansibleLimit, ansibleSkipTags, backoffLimit, etc.) as a JSON object\n2. Use legacy individual parameters: nodeSets (required array) and servicesOverride (optional array)\n\nThe spec must include 'nodeSets' field with at least one nodeSet name. Common spec fields include:\n- nodeSets: array of nodeSet names (required)\n- servicesOverride: array of service names to deploy\n- ansibleTags: string or array of Ansible tags to run\n- ansibleLimit: string to limit deployment to specific hosts\n- ansibleSkipTags: string or array of Ansible tags to skip\n- backoffLimit: number of retries before marking deployment failed\n- Any other valid OpenStackDataplaneDeployment spec fields"),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace where the OpenStackDataplaneDeployment CR will be created. Defaults to 'openstack' if not provided."),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the OpenStackDataplaneDeployment CR to create"),
		),
	)

	s.AddTool(createDataplaneDeploymentTool, handlers.CreateDataplaneDeploymentHandler(k8sClient))

	// Register the get_dataplane_deployment tool
	getDataplaneDeploymentTool := mcp.NewTool("get_dataplane_deployment",
		mcp.WithDescription("Query OpenStackDataplaneDeployment CRD to get spec and status information including nodeSets, servicesOverride, and deployment conditions."),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace where the OpenStackDataplaneDeployment CR is located. Defaults to 'openstack' if not provided."),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Name of the OpenStackDataplaneDeployment CR to query"),
		),
	)

	s.AddTool(getDataplaneDeploymentTool, handlers.GetDataplaneDeploymentHandler(k8sClient))

	// Register the list_dataplane_deployments tool
	listDataplaneDeploymentsTool := mcp.NewTool("list_dataplane_deployments",
		mcp.WithDescription("List all OpenStackDataplaneDeployment CRs in the specified namespace and return their spec and status information."),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace where the OpenStackDataplaneDeployment CRs are located. Defaults to 'openstack' if not provided."),
		),
	)

	s.AddTool(listDataplaneDeploymentsTool, handlers.ListDataplaneDeploymentsHandler(k8sClient))

	// Register the list_dataplane_nodesets tool
	listDataplaneNodeSetsTool := mcp.NewTool("list_dataplane_nodesets",
		mcp.WithDescription("List all OpenStackDataplaneNodeSet CRs in the specified namespace and return their spec and status information."),
		mcp.WithString("namespace",
			mcp.Description("Kubernetes namespace where the OpenStackDataplaneNodeSet CRs are located. Defaults to 'openstack' if not provided."),
		),
	)

	s.AddTool(listDataplaneNodeSetsTool, handlers.ListDataplaneNodeSetsHandler(k8sClient))

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
