resource "maas_dns_record" "test_a" {
  type = "A/AAAA"
  data = "10.99.11.33"
  fqdn = "test-a.${maas_dns_domain.cloudbase.name}"
}

resource "maas_dns_record" "test_aaaa" {
  type = "A/AAAA"
  data = "2001:db8:3333:4444:5555:6666:7777:8888"
  fqdn = "test-aaaa.${maas_dns_domain.cloudbase.name}"
}

resource "maas_dns_record" "test_txt" {
  type   = "TXT"
  data   = "@"
  name   = "test-txt"
  domain = maas_dns_domain.cloudbase.name
}
