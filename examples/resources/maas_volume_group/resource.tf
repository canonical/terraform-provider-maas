resource "maas_block_device" "vdb" {
  machine        = maas_machine.virsh_vm2.id
  name           = "vdb"
  id_path        = "/dev/vdb"
  size_gigabytes = 27
  tags = [
    "ssd",
  ]
}

resource "maas_volume_group" "vg1" {
  name          = "volume group 1"
  machine       = maas_machine.virsh_vm2.id
  block_devices = [maas_block_device.vdb.id]
}
