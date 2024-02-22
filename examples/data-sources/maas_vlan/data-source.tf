data "maas_vlan" "default" {
  fabric = data.maas_fabric.default.id
  vlan   = 0
}

data "maas_vlan" "vid10" {
  fabric = data.maas_fabric.default.id
  vlan   = 10
}
