// Test A record type
resource "porkbun_dns_record" "www" {
  domain  = var.domain
  name    = "www"
  type    = "A"
  content = "192.0.2.1"
  ttl     = "600"
}

// Test MX record type
resource "porkbun_dns_record" "mx_mail" {
  domain  = var.domain
  name    = "mail"
  type    = "MX"
  content = var.mx_server
  ttl     = "3600"
  prio    = "10"
}

// Test MX record type on root domain
resource "porkbun_dns_record" "root_mail" {
  domain  = var.domain
  name    = ""
  type    = "MX"
  content = var.mx_server
  ttl     = "3600"
  prio    = "10"
}
