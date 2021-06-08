data "maas_fabric" "default" {
  name = "maas"
}

data "maas_vlan" "default" {
  fabric_id = data.maas_fabric.default.id
  vid = 0
}

data "maas_vlan" "vid10" {
  fabric_id = data.maas_fabric.default.id
  vid = 10
}

data "maas_subnet" "pxe" {
  cidr = "10.99.0.0/16"
  vlan_id = data.maas_vlan.default.id
}

data "maas_subnet" "vid10" {
  cidr = "10.10.0.0/16"
  vlan_id = data.maas_vlan.vid10.id
}
