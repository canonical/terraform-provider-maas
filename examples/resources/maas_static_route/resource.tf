resource "maas_fabric" "tf_fabric" {
  name = "tf-fabric"
}

resource "maas_vlan" "tf_vlan" {
  fabric = maas_fabric.tf_fabric.id
  vid    = 14
  name   = "tf-vlan14"
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
}

resource "maas_static_route" "tf_static_route" {
  source      = maas_subnet.tf_subnet.name
  destination = "172.18.0.0/16"
  gateway_ip  = maas_subnet.tf_subnet.gateway_ip
  metric      = 55
}
