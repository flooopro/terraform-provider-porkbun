# porkbun_glue_record

Manages a Glue Record at Porkbun. Glue records are required when the nameservers for a domain are subdomains of the domain itself (e.g., nameserver `ns1.example.com` for domain `example.com`).

## Example Usage

This example shows how to set up self-hosted nameservers.

```hcl
variable "my_domain" {
  type    = string
  default = "example.com"
}

// Step 1: Create the glue records, which are A/AAAA records
// registered at the domain's registrar.
resource "porkbun_glue_record" "ns1" {
  domain = var.my_domain
  host   = "ns1"
  ips    = ["192.0.2.1", "2001:db8::1"]
}

resource "porkbun_glue_record" "ns2" {
  domain = var.my_domain
  host   = "ns2"
  ips    = ["192.0.2.2"]
}

// Step 2: Assign the nameservers to the domain.
// Use `depends_on` to ensure the glue records exist before assignment.
resource "porkbun_domain_nameservers" "custom_ns" {
  domain = var.my_domain
  nameservers = [
    "ns1.${var.my_domain}",
    "ns2.${var.my_domain}",
  ]

  depends_on = [
    porkbun_glue_record.ns1,
    porkbun_glue_record.ns2,
  ]
}
```

## Argument Reference

*   `domain` - (String, Required) The top-level domain for which the glue record is being created. Changing this forces a new resource.
*   `host` - (String, Required) The host part of the nameserver (e.g., `ns1` for `ns1.example.com`). Changing this forces a new resource.
*   `ips` - (List of Strings, Required) A list of IP addresses (IPv4 or IPv6) for the nameserver host.

## Attribute Reference

*   `id` - (String) A unique identifier for the glue record in the format `domain/host`.

## Import

You can import an existing glue record using the `domain/host` format.

```bash
terraform import porkbun_glue_record.ns1 example.com/ns1
```