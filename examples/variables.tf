variable "domain" {
  type = string
  default = "example.com"
}

variable "domain_ns_test" {
  type = string
  default = "example.com"
}

variable "mx_server" {
  type = string
  default = "mail.example.com"
}

variable "nameservers" {
  type = list(string)
  default = [
    "ns.example.com",
    "ns.example.com"
  ]
}

variable "glue_record" {
  type = list(string)
  default = [
    "10.0.0.1", 
    "2a03:ae0:da3:de::1"
    ]
}