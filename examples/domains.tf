data "porkbun_domains" "all" {}

output "all_my_domain_names" {
  value       = [for d in data.porkbun_domains.all.domains : d.domain]
}

output "all_my_domain_objects" {
  value       = data.porkbun_domains.all.domains
}
