# porkbun_dns_records (Data Source)

Provides a list of all DNS records for a specific domain.

## Example Usage

```hcl
data "porkbun_dns_records" "all_records_for_domain" {
  domain = "example.com"
}

output "all_a_records" {
  value = [for r in data.porkbun_dns_records.all_records_for_domain.records : r if r.type == "A"]
}
```

## Argument Reference

*   `domain` - (String, Required) The domain name for which to retrieve the DNS records.

## Attribute Reference

*   `id` - (String) The domain name.
*   `records` - (List of Objects) A list of all DNS records for the domain, with the following attributes for each:
    *   `id` - (String) The ID of the record.
    *   `name` - (String) The subdomain part of the record.
    *   `type` - (String) The type of the record.
    *   `content` - (String) The content/value of the record.
    *   `ttl` - (String) The TTL of the record.
    *   `prio` - (String) The priority of the record (for MX/SRV records).
    