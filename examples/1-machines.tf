#
# Machine 1
#
resource "maas_machine" "virsh_vm1" {
  power_type = "virsh"
  power_parameters = jsonencode({
    power_address = "qemu+ssh://ubuntu@10.113.1.26/system"
    power_id = "test-vm1"
  })
  pxe_mac_address = "52:54:00:89:f5:3e"
}

resource "maas_network_interface_physical" "virsh_vm1_nic1" {
  machine = maas_machine.virsh_vm1.id
  mac_address = "52:54:00:89:f5:3e"
  name = "eth0"
  vlan = data.maas_vlan.default.id
  tags = [
    "nic1-tag1",
    "nic1-tag2",
    "nic1-tag3",
  ]
}

resource "maas_network_interface_link" "virsh_vm1_nic1" {
  machine = maas_machine.virsh_vm1.id
  network_interface = maas_network_interface_physical.virsh_vm1_nic1.id
  subnet = data.maas_subnet.pxe.id
  mode = "STATIC"
  ip_address = "10.99.4.111"
  default_gateway = true
}

resource "maas_network_interface_physical" "virsh_vm1_nic2" {
  machine = maas_machine.virsh_vm1.id
  mac_address = "52:54:00:f5:89:ae"
  name = "eth1"
  vlan = data.maas_vlan.vid10.id
  tags = [
    "nic2-tag1",
    "nic2-tag2",
    "nic2-tag3",
  ]
}

resource "maas_network_interface_link" "virsh_vm1_nic2" {
  machine = maas_machine.virsh_vm1.id
  network_interface = maas_network_interface_physical.virsh_vm1_nic2.id
  subnet = data.maas_subnet.vid10.id
  mode = "AUTO"
}

resource "maas_network_interface_physical" "virsh_vm1_nic3" {
  machine = maas_machine.virsh_vm1.id
  mac_address = "52:54:00:0e:92:79"
  name = "eth2"
  vlan = data.maas_vlan.default.id
  tags = [
    "nic3-tag1",
    "nic3-tag2",
    "nic3-tag3",
  ]
}

resource "maas_network_interface_link" "virsh_vm1_nic3" {
  machine = maas_machine.virsh_vm1.id
  network_interface = maas_network_interface_physical.virsh_vm1_nic3.id
  subnet = data.maas_subnet.pxe.id
  mode = "DHCP"
}

#
# Machine 2
#
resource "maas_machine" "virsh_vm2" {
  power_type = "virsh"
  power_parameters = jsonencode({
    power_address = "qemu+ssh://ubuntu@10.113.1.26/system"
    power_id = "test-vm2"
  })
  pxe_mac_address = "52:54:00:7c:f7:77"
}

resource "maas_network_interface_physical" "virsh_vm2_nic1" {
  machine = maas_machine.virsh_vm2.id
  mac_address = "52:54:00:7c:f7:77"
  name = "eno0"
  vlan = data.maas_vlan.default.id
  tags = [
    "nic1-tag1",
    "nic1-tag2",
    "nic1-tag3",
  ]
}

resource "maas_network_interface_link" "virsh_vm2_nic1" {
  machine = maas_machine.virsh_vm2.id
  network_interface = maas_network_interface_physical.virsh_vm2_nic1.id
  subnet = data.maas_subnet.pxe.id
  mode = "STATIC"
  ip_address = "10.99.4.112"
  default_gateway = true
}

resource "maas_network_interface_physical" "virsh_vm2_nic2" {
  machine = maas_machine.virsh_vm2.id
  mac_address = "52:54:00:82:5c:c1"
  name = "eno1"
  vlan = data.maas_vlan.default.id
  tags = [
    "nic2-tag1",
    "nic2-tag2",
    "nic2-tag3",
  ]
}

resource "maas_network_interface_link" "virsh_vm2_nic2" {
  machine = maas_machine.virsh_vm2.id
  network_interface = maas_network_interface_physical.virsh_vm2_nic2.id
  subnet = data.maas_subnet.pxe.id
  mode = "DHCP"
}

resource "maas_network_interface_physical" "virsh_vm2_nic3" {
  machine = maas_machine.virsh_vm2.id
  mac_address = "52:54:00:bb:6e:9f"
  name = "eno2"
  vlan = data.maas_vlan.vid10.id
  tags = [
    "nic3-tag1",
    "nic3-tag2",
    "nic3-tag3",
  ]
}

resource "maas_network_interface_link" "virsh_vm2_nic3" {
  machine = maas_machine.virsh_vm2.id
  network_interface = maas_network_interface_physical.virsh_vm2_nic3.id
  subnet = data.maas_subnet.vid10.id
  mode = "AUTO"
}

resource "maas_block_device" "vdb" {
  machine = maas_machine.virsh_vm2.id
  name = "vdb"
  id_path = "/dev/vdb"
  size_gigabytes = 27
  tags = [
    "ssd",
  ]

  partitions {
    size_gigabytes = 10
    fs_type = "ext4"
    label = "media"
    mount_point = "/media"
  }

  partitions {
    size_gigabytes = 15
    fs_type = "ext4"
    mount_point = "/storage"
  }
}

resource "maas_block_device" "vdc" {
  machine = maas_machine.virsh_vm2.id
  name = "vdc"
  id_path = "/dev/vdc"
  size_gigabytes = 33

  partitions {
    size_gigabytes = 11
  }

  partitions {
    size_gigabytes = 13
    fs_type = "ext4"
    label = "images"
    mount_point = "/images"
  }
}

#
# Machine 3
#
resource "maas_machine" "virsh_vm3" {
  power_type = "virsh"
  power_parameters = jsonencode({
    power_address = "qemu+ssh://ubuntu@10.113.1.21/system"
    power_id = "machine-01"
  })
  pxe_mac_address = "52:54:00:16:78:ec"
}

#
# Machine 4
#
resource "maas_machine" "virsh_vm4" {
  power_type = "virsh"
  power_parameters = jsonencode({
    power_address = "qemu+ssh://ubuntu@10.113.1.22/system"
    power_id = "machine-05"
  })
  pxe_mac_address = "52:54:00:c4:74:96"
}
