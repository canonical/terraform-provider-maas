#
# Machine 1
#
resource "maas_machine" "virsh_vm1" {
  power_type = "virsh"
  power_parameters = {
    power_address = "qemu+ssh://ubuntu@10.113.1.26/system"
    power_id = "test-vm1"
  }
  pxe_mac_address = "52:54:00:89:f5:3e"
}

resource "maas_network_interface_physical" "virsh_vm1_nic1" {
  machine_id = maas_machine.virsh_vm1.id
  mac_address = "52:54:00:89:f5:3e"
  name = "eth0"
  vlan = data.maas_vlan.default.id
  tags = [
    "nic1-tag1",
    "nic1-tag2",
    "nic1-tag3",
  ]
}

resource "maas_network_interface_physical" "virsh_vm1_nic2" {
  machine_id = maas_machine.virsh_vm1.id
  mac_address = "52:54:00:f5:89:ae"
  name = "eth1"
  vlan = data.maas_vlan.vid10.id
  tags = [
    "nic2-tag1",
    "nic2-tag2",
    "nic2-tag3",
  ]
}

resource "maas_network_interface_physical" "virsh_vm1_nic3" {
  machine_id = maas_machine.virsh_vm1.id
  mac_address = "52:54:00:0e:92:79"
  name = "eth2"
  vlan = data.maas_vlan.default.id
  tags = [
    "nic3-tag1",
    "nic3-tag2",
    "nic3-tag3",
  ]
}

resource "maas_network_interface_link" "virsh_vm1_nic1" {
  machine_id = maas_machine.virsh_vm1.id
  network_interface_id = maas_network_interface_physical.virsh_vm1_nic1.id
  subnet_id = data.maas_subnet.pxe.id
  mode = "STATIC"
  ip_address = "10.113.1.111"
  default_gateway = true
}

resource "maas_network_interface_link" "virsh_vm1_nic2" {
  machine_id = maas_machine.virsh_vm1.id
  network_interface_id = maas_network_interface_physical.virsh_vm1_nic2.id
  subnet_id = data.maas_subnet.vid10.id
  mode = "AUTO"
}

resource "maas_network_interface_link" "virsh_vm1_nic3" {
  machine_id = maas_machine.virsh_vm1.id
  network_interface_id = maas_network_interface_physical.virsh_vm1_nic3.id
  subnet_id = data.maas_subnet.pxe.id
  mode = "DHCP"
}

#
# Machine 2
#
resource "maas_machine" "virsh_vm2" {
  power_type = "virsh"
  power_parameters = {
    power_address = "qemu+ssh://ubuntu@10.113.1.26/system"
    power_id = "test-vm2"
  }
  pxe_mac_address = "52:54:00:7c:f7:77"
}

resource "maas_network_interface_physical" "virsh_vm2_nic1" {
  machine_id = maas_machine.virsh_vm2.id
  mac_address = "52:54:00:7c:f7:77"
  name = "eno0"
  vlan = data.maas_vlan.default.id
  tags = [
    "nic1-tag1",
    "nic1-tag2",
    "nic1-tag3",
  ]
}

resource "maas_network_interface_physical" "virsh_vm2_nic2" {
  machine_id = maas_machine.virsh_vm2.id
  mac_address = "52:54:00:82:5c:c1"
  name = "eno1"
  vlan = data.maas_vlan.default.id
  tags = [
    "nic2-tag1",
    "nic2-tag2",
    "nic2-tag3",
  ]
}

resource "maas_network_interface_physical" "virsh_vm2_nic3" {
  machine_id = maas_machine.virsh_vm2.id
  mac_address = "52:54:00:bb:6e:9f"
  name = "eno2"
  vlan = data.maas_vlan.vid10.id
  tags = [
    "nic3-tag1",
    "nic3-tag2",
    "nic3-tag3",
  ]
}

resource "maas_network_interface_link" "virsh_vm2_nic1" {
  machine_id = maas_machine.virsh_vm2.id
  network_interface_id = maas_network_interface_physical.virsh_vm2_nic1.id
  subnet_id = data.maas_subnet.pxe.id
  mode = "STATIC"
  ip_address = "10.113.1.112"
  default_gateway = true
}

resource "maas_network_interface_link" "virsh_vm2_nic2" {
  machine_id = maas_machine.virsh_vm2.id
  network_interface_id = maas_network_interface_physical.virsh_vm2_nic2.id
  subnet_id = data.maas_subnet.vid10.id
  mode = "DHCP"
}

resource "maas_network_interface_link" "virsh_vm2_nic3" {
  machine_id = maas_machine.virsh_vm2.id
  network_interface_id = maas_network_interface_physical.virsh_vm2_nic3.id
  subnet_id = data.maas_subnet.pxe.id
  mode = "AUTO"
}
