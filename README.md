# Porkbun Provider for Terraform

The Porkbun provider for Terraform allows you to manage your domains, DNS records, and other resources on [Porkbun](https://porkbun.com) using Infrastructure as Code.

This provider is maintained by [flooopro].

## Example Usage

```hcl
terraform {
  required_providers {
    porkbun = {
      # The source path should match what you use for publishing
      source  = "your-namespace/porkbun"
      version = "1.0.0"
    }
  }
}

# Configure the Porkbun Provider
provider "porkbun" {
  # It is recommended to use environment variables for keys
}

# Create a simple A record
resource "porkbun_dns_record" "example" {
  domain  = "your-domain.com"
  name    = "www"
  type    = "A"
  content = "192.0.2.1"
}
```

## Authentication

The Porkbun provider requires an API Key and a Secret API Key to authenticate with the Porkbun API. You can generate these from the [API Access](https://porkbun.com/account/api) section of your Porkbun account.

There are two ways to provide these credentials to the provider:

### 1. Static Credentials (Not Recommended)

You can provide the credentials directly in the provider configuration block.

```hcl
provider "porkbun" {
  api_key        = "pk1_..."
  secret_api_key = "sk1_..."
}
```

### 2. Environment Variables (Recommended)

This is the recommended approach as it keeps your secret credentials out of your source code. Export the following environment variables:

```bash
export PORKBUN_API_KEY="pk1_..."
export PORKBUN_SECRET_API_KEY="sk1_..."
```

The provider will automatically use these variables if they are present.

## Schema

### Provider Arguments

*   `api_key` (String, Optional) - Your Porkbun API Key. Can also be provided via the `PORKBUN_API_KEY` environment variable.
*   `secret_api_key` (String, Optional) - Your Porkbun Secret API Key. Can also be provided via the `PORKBUN_SECRET_API_KEY` environment variable. Sensitive.