resource "maas_network_interface_vlan" "example" {
  machine = maas_machine.example.id
  parent  = maas_network_interface_bond.example.name
  vlan    = data.maas_vlan.example.id
  fabric  = "fabric"
}
