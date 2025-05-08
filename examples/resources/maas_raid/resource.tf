resource "maas_machine" "virsh_vm1" {
  power_type = "virsh"
  power_parameters = jsonencode({
    power_address = "qemu+ssh://ubuntu@10.113.1.26/system"
    power_id      = "test-vm1"
  })
  pxe_mac_address = "52:54:00:89:f5:3e"
}

resource "maas_block_device" "raidbd1" {
  machine        = maas_machine.virsh_vm1.id
  name           = "raidbd1"
  id_path        = "/dev/raidbd1"
  size_gigabytes = 27
  tags = [
    "ssd",
  ]
}

resource "maas_block_device" "raiddb2" {
  machine        = maas_machine.virsh_vm1.id
  name           = "raiddb2"
  id_path        = "/dev/raiddb2"
  size_gigabytes = 35
  tags = [
    "ssd",
  ]

  partitions {
    size_gigabytes = 30
  }
}

resource "maas_block_device" "raidbd3" {
  machine        = maas_machine.virsh_vm1.id
  name           = "raidbd3"
  id_path        = "/dev/raidbd3"
  size_gigabytes = 21
  tags = [
    "ssd",
  ]
}

resource "maas_raid" "raid1" {
  machine = maas_machine.virsh_vm1.id
  fs_type = "ext4"
  name    = "RAID 1"
  level   = "1"

  block_devices = [
    resource.maas_block_device.raidbd1.id
  ]
  partitions = [
    resource.maas_block_device.raiddb2.partitions.0.id
  ]
  spare_devices = [
    resource.maas_block_device.raidbd3.id
  ]
}
