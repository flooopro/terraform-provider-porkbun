# porkbun_domains (Data Source)

Provides a list of all domains in your Porkbun account.

## Example Usage

```hcl
# Get all domains
data "porkbun_domains" "all" {}

# Output a simple list of all domain names
output "all_my_domain_names" {
  value = [for d in data.porkbun_domains.all.domains : d.domain]
}

# Create a verification record for every domain in the account
resource "porkbun_dns_record" "verify_ownership" {
  for_each = { for d in data.porkbun_domains.all.domains : d.domain => d }

  domain  = each.value.domain
  name    = "_ownership-verification"
  type    = "A"
  content = "192.0.2.42"
}
```

## Attribute Reference

*   `id` - (String) A unique identifier for the data source call.
*   `domains` - (List of Objects) A list of all domains in the account, with the following attributes for each:
    *   `domain` - (String) The domain name.
    *   `status` - (String) The current status of the domain.
    *   `tld` - (String) The Top-Level Domain of the domain.
    *   `create_date` - (String) The creation date of the domain.
    *   `expire_date` - (String) The expiration date of the domain.
    *   `security_lock` - (Boolean) Whether the domain has a registrar lock enabled.
    *   `whois_privacy` - (Boolean) Whether WHOIS privacy is enabled for the domain.
    *   `auto_renew` - (Boolean) Whether auto-renewal is enabled for the domain.