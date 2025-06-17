# porkbun_dns_record

Manages a DNS Record on Porkbun.

## Example Usage

```hcl
resource "porkbun_dns_record" "mx_example" {
  domain  = "example.com"
  name    = "" // Use an empty string for the root domain
  type    = "MX"
  content = "mail.example.com."
  prio    = "10"
  ttl     = "3600"
}

resource "porkbun_dns_record" "txt_example" {
  domain  = "example.com"
  name    = "_dmarc"
  type    = "TXT"
  content = "\"v=DMARC1; p=none;\""
}
```

## Argument Reference

*   `domain` - (String, Required) The domain name for the record. Changing this forces a new resource to be created.
*   `name` - (String, Optional) The subdomain for the record. Use an empty string (`""`) for the root domain.
*   `type` - (String, Required) The type of the DNS record (e.g., `A`, `CNAME`, `TXT`, `MX`).
*   `content` - (String, Required) The content/value of the DNS record.
*   `ttl` - (String, Optional) The Time To Live (TTL) of the record in seconds. Defaults to `300`.
*   `prio` - (String, Optional) The priority of the record (for `MX` and `SRV` records only).

## Attribute Reference

In addition to the arguments above, the following attributes are exported:

*   `id` - (String) The unique ID of the DNS record, as assigned by Porkbun.

## Import

You can import an existing DNS record using the `domain/record_id` format.

```bash
terraform import porkbun_dns_record.example example.com/123456789
```