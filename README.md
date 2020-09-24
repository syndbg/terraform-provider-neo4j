# terraform-provider-neo4j

A Terraform provider for https://neo4j.com that manages users (for now).

**This is an alpha-quality provider**.

## Installation

1. Download latest GitHub release locally for your OS and architecture.
2. Follow https://www.terraform.io/docs/configuration/providers.html#third-party-plugins
3. Move downloaded release binary to local terraform plugin dir.

E.g installation on a Linux AMD64 host

```
# Assuming we already downloaded the binary at ~/Downloads/terraform-provider-neo4j_v0.1.0_linux_amd64
> mkdir -p ~/.terraform.d/plugins/linux_amd64
> mv ~/Downloads/terraform-provider-neo4j_v0.1.0_linux_amd64 ~/.terraform.d/plugins/linux_amd64/terraform-provider-neo4j_v0.1.0
```

## Usage

```hcl-terraform
provider "neo4j" {
  username = "neo4j"  
  password = "myneo4j"
  connection_uri = "neo4j://localhost:7687" 
  realm = ""
}

resource "neo4j_user" "user" {
  user = "myuser"
}
```
