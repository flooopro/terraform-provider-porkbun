resource "porkbun_domain_nameservers" "fenster_slowenien_ns" {
  domain = var.domain_ns_test
  nameservers = var.nameservers
}

resource "porkbun_glue_record" "ns1" {
  domain = var.domain_ns_test
  host   = "ns1"
  ips    = var.glue_record
}
