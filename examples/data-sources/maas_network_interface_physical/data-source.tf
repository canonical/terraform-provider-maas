resource "maas_network_interface_physical" "virsh_vm1_nic1" {
  machine     = maas_machine.virsh_vm1.id
  mac_address = "52:54:00:89:f5:3e"
  name        = "eth0"
  vlan        = data.maas_vlan.default.id
}

data "maas_network_interface_physical" "test_network_interface_physical" {
  machine = maas_network_interface_physical.virsh_vm1_nic1.machine
  name    = maas_network_interface_physical.virsh_vm1_nic1.name
}
