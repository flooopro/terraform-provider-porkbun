# porkbun_dnssec_record

Manages a DNSSEC DS (Delegation Signer) record for a domain. This is typically used to enable DNSSEC when using an external DNS provider.

## Example Usage

```hcl
# This example adds a DS record for a domain, as provided
# by an external DNS host like Cloudflare.

resource "porkbun_dnssec_record" "example_ds" {
  domain = "example.com"

  # Example values for Algorithm 13 (ECDSAP256SHA256)
  key_tag     = "2371"
  algorithm   = "13"
  digest_type = "2"
  digest      = "E2D3C916F6DEEAC73294E8268FB5885044A833FC5459588F4A9184CFC41A5766"
}
```

## Argument Reference

*   `domain` - (String, Required) The domain name. Changing this forces a new resource.
*   `key_tag` - (String, Required) The Key Tag of the DS record.
*   `algorithm` - (String, Required) The algorithm number used for the DS record.
*   `digest_type` - (String, Required) The digest type number used for the DS record.
*   `digest` - (String, Required) The digest (hash) of the DNSKEY record.

## Attribute Reference

*   `id` - (String) A unique, deterministic identifier for the record based on its content.

## Import

You can import an existing DNSSEC record using the format `domain/keyTag/algorithm/digestType/digest`.

```bash
terraform import porkbun_dnssec_record.example_ds example.com/2371/13/2/E2D3C916...
```