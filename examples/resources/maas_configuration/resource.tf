resource "maas_configuration" "test_bool" {
  key   = "release_notifications"
  value = "true"
}

resource "maas_configuration" "test_string" {
  key   = "default_osystem"
  value = "ubuntu"
}

resource "maas_configuration" "test_int" {
  key   = "maas_syslog_port"
  value = "5247"
}

resource "maas_configuration" "test_list" {
  key   = "dns_trusted_acl"
  value = "192.168.1.1 192.168.1.2"
}

resource "maas_configuration" "test_unset" {
  key   = "dns_trusted_acl"
  value = ""
}
