resource "maas_network_interface_link" "virsh_vm1_nic1" {
  machine           = maas_machine.virsh_vm1.id
  network_interface = maas_network_interface_physical.virsh_vm1_nic1.id
  subnet            = data.maas_subnet.pxe.id
  mode              = "STATIC"
  ip_address        = "10.99.4.111"
  default_gateway   = true
}

resource "maas_network_interface_link" "virsh_vm1_nic2" {
  machine           = maas_machine.virsh_vm1.id
  network_interface = maas_network_interface_physical.virsh_vm1_nic2.id
  subnet            = data.maas_subnet.vid10.id
  mode              = "AUTO"
}

resource "maas_network_interface_link" "virsh_vm1_nic3" {
  machine           = maas_machine.virsh_vm1.id
  network_interface = maas_network_interface_physical.virsh_vm1_nic3.id
  subnet            = data.maas_subnet.pxe.id
  mode              = "DHCP"
}
