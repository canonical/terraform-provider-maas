resource "maas_machine" "virsh_vm1" {
  power_type = "virsh"
  power_parameters = jsonencode({
    power_address = "qemu+ssh://ubuntu@10.113.1.26/system"
    power_id      = "test-vm1"
  })
  pxe_mac_address = "52:54:00:89:f5:3e"
}


resource "maas_block_device" "vdb1" {
  machine        = maas_machine.virsh_vm1.id
  name           = "vdb1"
  id_path        = "/dev/vdb1"
  size_gigabytes = 27
  tags = [
    "ssd",
  ]
}

resource "maas_block_device" "vdb2" {
  machine        = maas_machine.virsh_vm1.id
  name           = "vdb2"
  id_path        = "/dev/vdb2"
  size_gigabytes = 35
  tags = [
    "ssd",
  ]

  partitions {
    size_gigabytes = 30
  }
}

resource "maas_volume_group" "vg1" {
  name          = "volume group 1"
  machine       = maas_machine.virsh_vm1.id
  block_devices = [maas_block_device.vdb1.id]
  partitions    = [maas_block_device.vdb2.partitions.0.id]
}

resource "maas_logical_volume" "lvm1" {
  fs_type        = "ext4"
  machine        = maas_machine.virsh_vm1.id
  name           = "LVM 1"
  size_gigabytes = 50
  volume_group   = maas_volume_group.vg1.id
}
