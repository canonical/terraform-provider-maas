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

resource "maas_space" "tf_space" {
  name = "tf-space"
}

resource "maas_fabric" "tf_fabric" {
  name = "tf-fabric"
}

resource "maas_vlan" "tf_vlan" {
  fabric = maas_fabric.tf_fabric.id
  vid = 14
  name = "tf-vlan14"
  space = maas_space.tf_space.name
}

resource "maas_subnet" "tf_subnet" {
  cidr = "10.88.88.0/24"
  fabric = maas_fabric.tf_fabric.id
  vlan = maas_vlan.tf_vlan.vid
}
