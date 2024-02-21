resource "maas_network_interface_physical" "virsh_vm1_nic1" {
  machine     = maas_machine.virsh_vm1.id
  mac_address = "52:54:00:89:f5:3e"
  name        = "eth0"
  vlan        = data.maas_vlan.default.id
  tags = [
    "nic1-tag1",
    "nic1-tag2",
    "nic1-tag3",
  ]
}

resource "maas_network_interface_physical" "virsh_vm1_nic2" {
  machine     = maas_machine.virsh_vm1.id
  mac_address = "52:54:00:f5:89:ae"
  name        = "eth1"
  vlan        = data.maas_vlan.vid10.id
  tags = [
    "nic2-tag1",
    "nic2-tag2",
    "nic2-tag3",
  ]
}

resource "maas_network_interface_physical" "virsh_vm1_nic3" {
  machine     = maas_machine.virsh_vm1.id
  mac_address = "52:54:00:0e:92:79"
  name        = "eth2"
  vlan        = data.maas_vlan.default.id
  tags = [
    "nic3-tag1",
    "nic3-tag2",
    "nic3-tag3",
  ]
}
