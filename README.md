# Terraform Provider for "The Cloud"

A Terraform provider for managing resources on **The Cloud** platform. This provider allows you to define your cloud infrastructure (VPCs, Instances, Load Balancers, etc.) as code using HashiCorp Configuration Language (HCL).

## 🚀 Quick Start

### Installation

To use this provider, add it to your `main.tf` file:

```hcl
terraform {
  required_providers {
    thecloud = {
      source  = "poyrazk/thecloud"
      version = "~> 0.1.0"
    }
  }
}

provider "thecloud" {
  api_key  = var.thecloud_api_key
  endpoint = "https://api.thecloud.com" # Optional: defaults to production API
}
```

### Authentication

The provider requires an **API Key** from The Cloud dashboard. You can provide it via the `api_key` attribute in the provider block or via the `THECLOUD_API_KEY` environment variable.

## 📦 Supported Resources

| Category | Resource | Description |
|----------|----------|-------------|
| **Networking** | `thecloud_vpc` | Manage Virtual Private Clouds and CIDR blocks. |
| **Networking** | `thecloud_subnet` | Segment VPCs into smaller networks with specific CIDRs. |
| **Compute** | `thecloud_instance` | Launch and manage virtual machine instances. |
| **Compute** | `thecloud_scaling_group` | Auto-scaling groups for dynamic instance management. |
| **Storage** | `thecloud_volume` | Block storage volumes for persistent data. |
| **Storage** | `thecloud_snapshot` | Create point-in-time backups of volumes. |
| **Database** | `thecloud_database` | Managed database services (Postgres, MySQL, Redis). |
| **Security** | `thecloud_security_group` | Virtual firewalls for controlling traffic. |
| **Security** | `thecloud_security_group_rule` | Specific ingress/egress rules for security groups. |
| **Secrets** | `thecloud_secret` | Securely store and manage sensitive information. |
| **Identity** | `thecloud_api_key` | Manage additional API keys via Terraform. |
| **Traffic** | `thecloud_load_balancer` | Highly available load balancing as a service. |
| **Traffic** | `thecloud_load_balancer_target` | Register instances to a load balancer. |

## ✨ Advanced Features

### 🛡️ API Resilience
The provider is built with high reliability in mind:
- **Automatic Retries**: Built-in exponential backoff for transient API errors (using `go-retryablehttp`).
- **Async Resource Management**: Intelligent polling for resources that require background provisioning (like Scaling Groups).

### ⏳ Configurable Timeouts
For long-running operations, you can customize wait times:
```hcl
resource "thecloud_instance" "large_node" {
  timeouts {
    create = "15m"
    delete = "10m"
  }
}
```

## 📖 Example Usage

```hcl
# Create a VPC
resource "thecloud_vpc" "main" {
  name       = "main-network"
  cidr_block = "10.0.0.0/16"
}

# Segment VPC into a Subnet
resource "thecloud_subnet" "public" {
  name              = "public-subnet"
  vpc_id            = thecloud_vpc.main.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = "us-east-1a"
}

# Create a Security Group
resource "thecloud_security_group" "web" {
  name        = "web-sg"
  description = "Public access for web servers"
  vpc_id      = thecloud_vpc.main.id
}

# Add Ingress Rule
resource "thecloud_security_group_rule" "http" {
  security_group_id = thecloud_security_group.web.id
  direction         = "ingress"
  protocol          = "tcp"
  port_min          = 80
  port_max          = 80
  cidr              = "0.0.0.0/0"
}

# Launch an Instance in the Subnet
resource "thecloud_instance" "web_server" {
  name      = "frontend-01"
  image     = "ubuntu-22.04"
  vpc_id    = thecloud_vpc.main.id
  subnet_id = thecloud_subnet.public.id
  ports     = "80:80,443:443"
}

# Provision a Managed Database
resource "thecloud_database" "products" {
  name    = "product-db"
  engine  = "postgres"
  version = "15"
  vpc_id  = thecloud_vpc.main.id
}
```

## 🛠️ Development

### Prerequisites
- [Go](https://golang.org/doc/install) 1.21+
- [Terraform](https://developer.hashicorp.com/terraform/downloads) 1.0+
- [Make](https://www.gnu.org/software/make/)

### Local Installation
To build the provider and install it in your local terraform plugins directory:
```bash
make install
```

### Running Tests
```bash
# Unit tests
make test

# Acceptance tests (requires a running API)
export THECLOUD_API_KEY="your-test-key"
export THECLOUD_ENDPOINT="http://localhost:8080"
make testacc
```

## 📄 License
MIT License. See [LICENSE](LICENSE) for more details.
