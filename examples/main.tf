terraform {
  required_providers {
    porkbun = {
      source  = "github.com/flooopro/porkbun"
      version = "0.0.1"
    }
  }
}

provider "porkbun" {}

output "www_record_id" {
  value = porkbun_dns_record.www.id
}

