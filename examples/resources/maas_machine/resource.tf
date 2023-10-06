resource "maas_machine" "virsh_vm1" {
  power_type = "virsh"
  power_parameters = jsonencode({
    power_address = "qemu+ssh://ubuntu@10.113.1.26/system"
    power_id = "test-vm1"
  })
  pxe_mac_address = "52:54:00:89:f5:3e"
}
