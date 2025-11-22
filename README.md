# OpenStack K8s MCP

A Model Context Protocol (MCP) server for querying OpenStack Kubernetes operators. This tool provides MCP-compatible access to OpenStackVersion CRDs in your Kubernetes cluster.

## Features

- **get_openstack_version**: Query OpenStackVersion CRD to retrieve:
  - Target version
  - Available version
  - Deployed version
  - Status conditions

- **set_openstack_version_targetversion**: Patch the targetVersion field of an OpenStackVersion CRD to set a new target version

## Prerequisites

- Go 1.22 or higher
- Access to a Kubernetes cluster with OpenStack operators installed
- Valid kubeconfig or in-cluster configuration

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd openstack-k8s-mcp

# Build the binary
go build -o openstack-k8s-mcp .
```

## Usage

### Running the MCP Server

The server communicates over stdio following the MCP protocol:

```bash
./openstack-k8s-mcp
```

### MCP Tool: get\_openstack\_version

Query an OpenStackVersion custom resource:

**Parameters:**
- `namespace` (required): Kubernetes namespace where the OpenStackVersion CR is located
- `name` (required): Name of the OpenStackVersion CR to query

**Returns:**
JSON object containing:
- `name`: CR name
- `namespace`: CR namespace
- `spec.targetVersion`: Desired OpenStack version
- `status.availableVersion`: Available version
- `status.deployedVersion`: Currently deployed version
- `status.conditions`: Array of condition objects with type, status, reason, message, and lastTransitionTime

### Example Response

```json
{
  "name": "openstack",
  "namespace": "openstack",
  "spec": {
    "targetVersion": "0.3.0"
  },
  "status": {
    "availableVersion": "0.3.0",
    "deployedVersion": "0.3.0",
    "conditions": [
      {
        "type": "Ready",
        "status": "True",
        "lastTransitionTime": "2025-01-15T10:30:00Z",
        "reason": "AllComponentsReady",
        "message": "All OpenStack components are ready"
      }
    ]
  }
}
```

### MCP Tool: set_openstack_version_targetversion

Patch the targetVersion field of an OpenStackVersion custom resource:

**Parameters:**
- `namespace` (optional): Kubernetes namespace where the OpenStackVersion CR is located. Defaults to `openstack` if not provided.
- `name` (required): Name of the OpenStackVersion CR to patch
- `targetVersion` (required): The target version to set for the OpenStackVersion CR

**Returns:**
JSON object containing:
- `name`: CR name
- `namespace`: CR namespace
- `spec.targetVersion`: Updated target OpenStack version
- `status.availableVersion`: Available version
- `status.deployedVersion`: Currently deployed version

### Example Response

```json
{
  "name": "openstack",
  "namespace": "openstack",
  "spec": {
    "targetVersion": "0.4.0"
  },
  "status": {
    "availableVersion": "0.3.0",
    "deployedVersion": "0.3.0"
  }
}
```

## Configuration with Claude Desktop

Add to your Claude Desktop MCP settings:

```json
{
  "mcpServers": {
    "openstack-k8s": {
      "command": "/path/to/openstack-k8s-mcp"
    }
  }
}
```

## Development

### Project Structure

- `main.go`: MCP server implementation
- `client.go`: Kubernetes client wrapper
- `go.mod`: Go module dependencies

### Dependencies

- `github.com/mark3labs/mcp-go`: MCP SDK for Go
- `github.com/openstack-k8s-operators/openstack-operator/apis`: OpenStack operator API types
- `k8s.io/client-go`: Kubernetes Go client
- `k8s.io/apimachinery`: Kubernetes API machinery

## License

MIT
