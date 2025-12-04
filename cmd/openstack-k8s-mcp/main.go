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
		mcp.WithDescription("Get OpenStack version information including targetVersion, availableVersion, deployedVersion, and conditions."),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: openstack)"),
		),
		mcp.WithString("name",
			mcp.Description("OpenStackVersion CR name (optional, auto-discovers if not provided)"),
		),
	)

	s.AddTool(getOpenStackVersionTool, handlers.GetOpenStackVersionHandler(k8sClient))

	// Register the update_openstack_version tool
	updateOpenStackVersionTool := mcp.NewTool("update_openstack_version",
		mcp.WithDescription("Update the targetVersion of the first OpenStackVersion CR in the namespace."),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: openstack)"),
		),
		mcp.WithString("targetVersion",
			mcp.Required(),
			mcp.Description("Target version to set (e.g., '0.0.2')"),
		),
	)

	s.AddTool(updateOpenStackVersionTool, handlers.UpdateOpenStackVersionHandler(k8sClient))

	// Register the wait_openstack_version tool
	waitOpenStackVersionTool := mcp.NewTool("wait_openstack_version",
		mcp.WithDescription("Wait for a condition on OpenStackVersion CR to become True. Common conditions: MinorUpdateReady, Ready, DeploymentReady, Available."),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: openstack)"),
		),
		mcp.WithString("name",
			mcp.Description("OpenStackVersion CR name (optional, auto-discovers if not provided)"),
		),
		mcp.WithString("condition",
			mcp.Required(),
			mcp.Description("Condition type to wait for (e.g., 'MinorUpdateReady', 'Ready')"),
		),
		mcp.WithNumber("timeout",
			mcp.Description("Timeout in seconds (default: 300)"),
		),
		mcp.WithNumber("pollInterval",
			mcp.Description("Poll interval in seconds (default: 5)"),
		),
	)

	s.AddTool(waitOpenStackVersionTool, handlers.WaitOpenStackVersionHandler(k8sClient))

	// Register the get_openstack_controlplane tool
	getOpenStackControlPlaneTool := mcp.NewTool("get_openstack_controlplane",
		mcp.WithDescription("Get OpenStackControlPlane spec and status including service configurations and conditions."),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: openstack)"),
		),
		mcp.WithString("name",
			mcp.Description("OpenStackControlPlane CR name (optional, auto-discovers if not provided)"),
		),
	)

	s.AddTool(getOpenStackControlPlaneTool, handlers.GetOpenStackControlPlaneHandler(k8sClient))

	// Register the verify_openstack_controlplane tool
	verifyOpenStackControlPlaneTool := mcp.NewTool("verify_openstack_controlplane",
		mcp.WithDescription("Verify all conditions on OpenStackControlPlane CR are ready. Returns allReady status and lists of ready/not-ready conditions."),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: openstack)"),
		),
		mcp.WithString("name",
			mcp.Description("OpenStackControlPlane CR name (optional, auto-discovers if not provided)"),
		),
	)

	s.AddTool(verifyOpenStackControlPlaneTool, handlers.VerifyOpenStackControlPlaneHandler(k8sClient))

	// Register the create_dataplane_deployment tool
	createDataplaneDeploymentTool := mcp.NewTool("create_dataplane_deployment",
		mcp.WithDescription("Create OpenStackDataplaneDeployment CR. Requires nodeSets array. Optional: servicesOverride, ansibleTags, ansibleLimit."),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: openstack)"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Deployment CR name (use dashes/underscores, not dots)"),
		),
	)

	s.AddTool(createDataplaneDeploymentTool, handlers.CreateDataplaneDeploymentHandler(k8sClient))

	// Register the create_dataplane_deployment_ovn tool
	createDataplaneDeploymentOVNTool := mcp.NewTool("create_dataplane_deployment_ovn",
		mcp.WithDescription("Create OpenStackDataplaneDeployment CR with servicesOverride=['ovn']. Auto-discovers all nodeSets."),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: openstack)"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Deployment CR name (use dashes/underscores, not dots)"),
		),
	)

	s.AddTool(createDataplaneDeploymentOVNTool, handlers.CreateDataplaneDeploymentOVNHandler(k8sClient))

	// Register the create_dataplane_deployment_update tool
	createDataplaneDeploymentUpdateTool := mcp.NewTool("create_dataplane_deployment_update",
		mcp.WithDescription("Create OpenStackDataplaneDeployment CR with servicesOverride=['update']. Auto-discovers all nodeSets."),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: openstack)"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Deployment CR name (use dashes/underscores, not dots)"),
		),
	)

	s.AddTool(createDataplaneDeploymentUpdateTool, handlers.CreateDataplaneDeploymentUpdateHandler(k8sClient))

	// Register the get_dataplane_deployment tool
	getDataplaneDeploymentTool := mcp.NewTool("get_dataplane_deployment",
		mcp.WithDescription("Get OpenStackDataplaneDeployment spec and status including nodeSets, conditions, and deployment statuses."),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: openstack)"),
		),
		mcp.WithString("name",
			mcp.Required(),
			mcp.Description("Deployment CR name"),
		),
	)

	s.AddTool(getDataplaneDeploymentTool, handlers.GetDataplaneDeploymentHandler(k8sClient))

	// Register the list_dataplane_deployments tool
	listDataplaneDeploymentsTool := mcp.NewTool("list_dataplane_deployments",
		mcp.WithDescription("List all OpenStackDataplaneDeployment CRs in namespace."),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: openstack)"),
		),
	)

	s.AddTool(listDataplaneDeploymentsTool, handlers.ListDataplaneDeploymentsHandler(k8sClient))

	// Register the list_dataplane_nodesets tool
	listDataplaneNodeSetsTool := mcp.NewTool("list_dataplane_nodesets",
		mcp.WithDescription("List all OpenStackDataplaneNodeSet CRs in namespace."),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: openstack)"),
		),
	)

	s.AddTool(listDataplaneNodeSetsTool, handlers.ListDataplaneNodeSetsHandler(k8sClient))

	// Register the verify_openstack_dataplanenodesets tool
	verifyDataplaneNodeSetsTool := mcp.NewTool("verify_openstack_dataplanenodesets",
		mcp.WithDescription("Verify all conditions on all OpenStackDataplaneNodeSet CRs are ready. Returns allReady status and lists of ready/not-ready NodeSets."),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: openstack)"),
		),
	)

	s.AddTool(verifyDataplaneNodeSetsTool, handlers.VerifyDataplaneNodeSetsHandler(k8sClient))

	// Register the get_resume_step tool
	getResumeStepTool := mcp.NewTool("get_resume_step",
		mcp.WithDescription("Determine which upgrade step to resume from based on current state. Analyzes targetVersion, availableVersion, and notReadyConditions to calculate the exact step number. Returns resumeStep (2-10) and explanation."),
		mcp.WithString("namespace",
			mcp.Description("Namespace (default: openstack)"),
		),
		mcp.WithString("name",
			mcp.Description("OpenStackVersion CR name (optional, auto-discovers if not provided)"),
		),
	)

	s.AddTool(getResumeStepTool, handlers.GetResumeStepHandler(k8sClient))

	// Start the server
	if err := server.ServeStdio(s); err != nil {
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
