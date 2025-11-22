# OpenStack K8s MCP

A Model Context Protocol (MCP) server for querying OpenStack Kubernetes operators. This tool provides MCP-compatible access to OpenStackVersion CRDs in your Kubernetes cluster.

## Features

- **get_openstack_version**: Query OpenStackVersion CRD to retrieve:
  - Target version
  - Available version
  - Deployed version
  - Status conditions

- **update_openstack_version**: Patch the targetVersion and optionally customContainerImages fields of an OpenStackVersion CRD

- **get_openstack_controlplane**: Query OpenStackControlPlane CRD to retrieve spec and status information

- **create_dataplane_deployment**: Create an OpenStackDataplaneDeployment CR to deploy services on dataplane nodes

- **get_dataplane_deployment**: Query OpenStackDataplaneDeployment CRD to retrieve spec and status information

- **list_dataplane_deployments**: List all OpenStackDataplaneDeployment CRs in a namespace

- **list_dataplane_nodesets**: List all OpenStackDataplaneNodeSet CRs in a namespace

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

### MCP Tool: update\_openstack\_version

Patch the targetVersion and optionally customContainerImages fields of an OpenStackVersion custom resource:

**Parameters:**
- `namespace` (optional): Kubernetes namespace where the OpenStackVersion CR is located. Defaults to `openstack` if not provided.
- `name` (required): Name of the OpenStackVersion CR to patch
- `targetVersion` (required): The target version to set for the OpenStackVersion CR
- `customContainerImages` (optional): Map of service names to custom container image URLs. If not provided, customContainerImages will not be modified.

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

### Example Usage with customContainerImages

```json
{
  "name": "openstack",
  "namespace": "openstack",
  "targetVersion": "0.4.0",
  "customContainerImages": {
    "nova": "quay.io/openstack-k8s-operators/nova-operator:v0.4.0",
    "neutron": "quay.io/openstack-k8s-operators/neutron-operator:v0.4.0"
  }
}
```

### MCP Tool: get\_openstack\_controlplane

Query an OpenStackControlPlane custom resource:

**Parameters:**
- `namespace` (optional): Kubernetes namespace where the OpenStackControlPlane CR is located. Defaults to `openstack` if not provided.
- `name` (optional): Name of the OpenStackControlPlane CR to query. If not provided, returns the first CR found in the namespace.

**Returns:**
JSON object containing:
- `name`: CR name
- `namespace`: CR namespace
- `spec`: ControlPlane specification including:
  - Service configurations (Nova, Neutron, Cinder, Glance, etc.)
  - Database configuration
  - Message queue configuration
  - Other controlplane settings
- `status`: ControlPlane status including:
  - `conditions`: Array of condition objects with deployment status
  - Service-specific status information
  - Other status fields as available

### Example Response

```json
{
  "name": "openstack",
  "namespace": "openstack",
  "spec": {
    "secret": "osp-secret",
    "storageClass": "local-storage",
    "nova": {
      "enabled": true,
      "template": {
        "secret": "osp-secret"
      }
    },
    "neutron": {
      "enabled": true,
      "template": {
        "secret": "osp-secret"
      }
    },
    "cinder": {
      "enabled": true,
      "template": {
        "cinderVolumes": {
          "volume1": {
            "storageClass": "local-storage"
          }
        }
      }
    }
  },
  "status": {
    "conditions": [
      {
        "type": "Ready",
        "status": "True",
        "lastTransitionTime": "2025-01-15T10:30:00Z",
        "reason": "AllServicesReady",
        "message": "All OpenStack services are ready"
      }
    ]
  }
}
```

### MCP Tool: create\_dataplane\_deployment

Create an OpenStackDataplaneDeployment custom resource to deploy services on dataplane nodes:

**Parameters:**
- `namespace` (optional): Kubernetes namespace where the OpenStackDataplaneDeployment CR will be created. Defaults to `openstack` if not provided.
- `name` (required): Name of the OpenStackDataplaneDeployment CR to create
- `nodeSets` (required): Array of nodeSet names to deploy to. Must contain at least one nodeSet.
- `servicesOverride` (optional): Array of service names to override the default services for the deployment

**Returns:**
Success message confirming the creation of the OpenStackDataplaneDeployment CR with the specified nodeSets and optional servicesOverride.

### Example Usage

Basic deployment to a single nodeSet:

```json
{
  "name": "my-deployment",
  "namespace": "openstack",
  "nodeSets": ["compute-nodes"]
}
```

Deployment to multiple nodeSets with service override:

```json
{
  "name": "my-deployment",
  "namespace": "openstack",
  "nodeSets": ["compute-nodes", "storage-nodes"],
  "servicesOverride": ["nova", "neutron", "ovn"]
}
```

### MCP Tool: get\_dataplane\_deployment

Query an OpenStackDataplaneDeployment custom resource to retrieve deployment information:

**Parameters:**
- `namespace` (optional): Kubernetes namespace where the OpenStackDataplaneDeployment CR is located. Defaults to `openstack` if not provided.
- `name` (required): Name of the OpenStackDataplaneDeployment CR to query

**Returns:**
JSON object containing:
- `name`: CR name
- `namespace`: CR namespace
- `spec`: Deployment specification including:
  - `nodeSets`: Array of nodeSet names
  - `servicesOverride`: Array of service names (if specified)
- `status`: Deployment status including:
  - `conditions`: Array of condition objects with deployment progress and status
  - `nodeSetConditions`: Per-nodeSet deployment status
  - Other status fields as available

### Example Response

```json
{
  "name": "my-deployment",
  "namespace": "openstack",
  "spec": {
    "nodeSets": ["compute-nodes"],
    "servicesOverride": ["nova", "neutron"]
  },
  "status": {
    "conditions": [
      {
        "type": "Ready",
        "status": "True",
        "lastTransitionTime": "2025-01-15T10:30:00Z",
        "reason": "DeploymentComplete",
        "message": "All nodeSets have been deployed successfully"
      }
    ],
    "nodeSetConditions": {
      "compute-nodes": {
        "Ready": {
          "status": "True",
          "message": "NodeSet deployment complete"
        }
      }
    }
  }
}
```

### MCP Tool: list\_dataplane\_deployments

List all OpenStackDataplaneDeployment custom resources in a namespace:

**Parameters:**
- `namespace` (optional): Kubernetes namespace where the OpenStackDataplaneDeployment CRs are located. Defaults to `openstack` if not provided.

**Returns:**
JSON array containing objects with:
- `name`: Deployment CR name
- `namespace`: Deployment CR namespace
- `spec`: Deployment specification including:
  - `nodeSets`: Array of nodeSet names
  - `servicesOverride`: Array of service names (if specified)
- `status`: Deployment status including:
  - `conditions`: Array of condition objects
  - `nodeSetConditions`: Per-nodeSet deployment status
  - Other status fields as available

### Example Response

```json
[
  {
    "name": "edpm-deployment",
    "namespace": "openstack",
    "spec": {
      "nodeSets": ["compute-nodes", "storage-nodes"],
      "servicesOverride": ["nova", "neutron", "ovn"]
    },
    "status": {
      "conditions": [
        {
          "type": "Ready",
          "status": "True",
          "lastTransitionTime": "2025-01-15T10:30:00Z",
          "reason": "DeploymentComplete",
          "message": "All nodeSets have been deployed successfully"
        }
      ],
      "nodeSetConditions": {
        "compute-nodes": {
          "Ready": {
            "status": "True",
            "message": "NodeSet deployment complete"
          }
        },
        "storage-nodes": {
          "Ready": {
            "status": "True",
            "message": "NodeSet deployment complete"
          }
        }
      }
    }
  },
  {
    "name": "upgrade-deployment",
    "namespace": "openstack",
    "spec": {
      "nodeSets": ["compute-nodes"]
    },
    "status": {
      "conditions": [
        {
          "type": "Ready",
          "status": "False",
          "lastTransitionTime": "2025-01-15T11:00:00Z",
          "reason": "DeploymentInProgress",
          "message": "Deployment in progress"
        }
      ]
    }
  }
]
```

### MCP Tool: list\_dataplane\_nodesets

List all OpenStackDataplaneNodeSet custom resources in a namespace:

**Parameters:**
- `namespace` (optional): Kubernetes namespace where the OpenStackDataplaneNodeSet CRs are located. Defaults to `openstack` if not provided.

**Returns:**
JSON array containing objects with:
- `name`: NodeSet CR name
- `namespace`: NodeSet CR namespace
- `spec`: NodeSet specification including:
  - `nodeTemplate`: Template for node configuration
  - `nodes`: Map of node names to their configurations
  - `services`: Services to deploy on the nodeSet
  - Other specification fields
- `status`: NodeSet status including:
  - `conditions`: Array of condition objects
  - `deploymentStatus`: Deployment progress information
  - Other status fields as available

### Example Response

```json
[
  {
    "name": "compute-nodes",
    "namespace": "openstack",
    "spec": {
      "nodeTemplate": {
        "ansibleSSHPrivateKeySecret": "dataplane-ansible-ssh-private-key-secret"
      },
      "nodes": {
        "compute-0": {
          "hostName": "compute-0",
          "ansible": {
            "ansibleHost": "192.168.1.10"
          }
        },
        "compute-1": {
          "hostName": "compute-1",
          "ansible": {
            "ansibleHost": "192.168.1.11"
          }
        }
      },
      "services": ["nova", "neutron", "ovn"]
    },
    "status": {
      "conditions": [
        {
          "type": "Ready",
          "status": "True",
          "lastTransitionTime": "2025-01-15T10:30:00Z",
          "message": "NodeSet is ready"
        }
      ]
    }
  },
  {
    "name": "storage-nodes",
    "namespace": "openstack",
    "spec": {
      "nodes": {
        "storage-0": {
          "hostName": "storage-0",
          "ansible": {
            "ansibleHost": "192.168.1.20"
          }
        }
      },
      "services": ["ceph"]
    },
    "status": {
      "conditions": [
        {
          "type": "Ready",
          "status": "True",
          "message": "NodeSet is ready"
        }
      ]
    }
  }
]
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
