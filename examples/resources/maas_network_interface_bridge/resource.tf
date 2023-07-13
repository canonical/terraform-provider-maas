resource "maas_network_interface_bridge" "example" {
  machine = maas_machine.example.id
  name    = "example"
  parent  = maas_network_interface_vlan.example.id
}
