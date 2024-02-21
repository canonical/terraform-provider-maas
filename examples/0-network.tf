#
# Spaces
#
resource "maas_space" "tf_space" {
  name = "tf-space"
}

#
# Fabrics
#
data "maas_fabric" "default" {
  name = "maas"
}

resource "maas_fabric" "tf_fabric" {
  name = "tf-fabric"
}

#
# VLANs
#
data "maas_vlan" "default" {
  fabric = data.maas_fabric.default.id
  vlan   = 0
}

data "maas_vlan" "vid10" {
  fabric = data.maas_fabric.default.id
  vlan   = 10
}

resource "maas_vlan" "tf_vlan" {
  fabric = maas_fabric.tf_fabric.id
  vid    = 14
  name   = "tf-vlan14"
  space  = maas_space.tf_space.name
}

#
# Subnets
#
data "maas_subnet" "pxe" {
  cidr = "10.99.0.0/16"
}

data "maas_subnet" "vid10" {
  cidr = "10.10.0.0/16"
}

resource "maas_subnet" "tf_subnet" {
  cidr       = "10.88.88.0/24"
  fabric     = maas_fabric.tf_fabric.id
  vlan       = maas_vlan.tf_vlan.vid
  name       = "tf_subnet"
  gateway_ip = "10.88.88.1"
  dns_servers = [
    "1.1.1.1",
  ]
  ip_ranges {
    type     = "reserved"
    start_ip = "10.88.88.1"
    end_ip   = "10.88.88.50"
  }
  ip_ranges {
    type     = "dynamic"
    start_ip = "10.88.88.200"
    end_ip   = "10.88.88.254"
  }
}

resource "maas_subnet" "tf_subnet_2" {
  cidr       = "10.77.77.0/24"
  name       = "tf_subnet_2"
  fabric     = maas_fabric.tf_fabric.id
  gateway_ip = "10.77.77.1"
  dns_servers = [
    "1.1.1.1",
  ]
}

resource "maas_subnet_ip_range" "dynamic_ip_range" {
  subnet   = maas_subnet.tf_subnet_2.id
  type     = "dynamic"
  start_ip = "10.77.77.2"
  end_ip   = "10.77.77.60"
}

resource "maas_subnet_ip_range" "reserved_ip_range" {
  subnet   = maas_subnet.tf_subnet_2.id
  type     = "reserved"
  start_ip = "10.77.77.200"
  end_ip   = "10.77.77.254"
  comment  = "Reserved for Static IPs"
}

#
# DNS Domains
#
resource "maas_dns_domain" "cloudbase" {
  name          = "cloudbase"
  ttl           = 3600
  authoritative = true
}

#
# DNS Records
#
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
