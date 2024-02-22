resource "maas_dns_domain" "cloudbase" {
  name          = "cloudbase"
  ttl           = 3600
  authoritative = true
}
