resource "maas_block_device" "vdb1" {
  machine        = maas_machine.virsh_vm2.id
  name           = "vdb1"
  id_path        = "/dev/vdb1"
  size_gigabytes = 27
  tags = [
    "ssd",
  ]
}

resource "maas_block_device" "vdb2" {
  machine        = maas_machine.virsh_vm2.id
  name           = "vdb2"
  id_path        = "/dev/vdb2"
  size_gigabytes = 35
  tags = [
    "ssd",
  ]

  partitions {
    size_gigabytes = 30
  }

  lifecycle {
    ignore_changes = [
      partitions[0].fs_type
    ]
  }
}

resource "maas_volume_group" "vg1" {
  name          = "volume group 1"
  machine       = maas_machine.virsh_vm2.id
  block_devices = [maas_block_device.vdb1.id]
  partitions    = [maas_block_device.vdb2.partitions.0.id]
}
