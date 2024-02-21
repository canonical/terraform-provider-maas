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
