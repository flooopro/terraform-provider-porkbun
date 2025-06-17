# porkbun_domain_nameservers

Manages the authoritative nameservers for a domain registered with Porkbun.

## Example Usage

```hcl
# Set custom nameservers for a domain
resource "porkbun_domain_nameservers" "example_ns" {
  domain = "example.com"
  nameservers = [
    "ns1.external-dns.com",
    "ns2.external-dns.com",
  ]
}
```

## Argument Reference

*   `domain` - (String, Required) The domain name whose nameservers will be managed. Changing this forces a new resource to be created.
*   `nameservers` - (List of Strings, Required) A list of the nameservers for the domain.

## Attribute Reference

*   `id` - (String) The domain name.

## Import

You can import existing nameserver configurations using the domain name.

```bash
terraform import porkbun_domain_nameservers.example_ns example.com
```
