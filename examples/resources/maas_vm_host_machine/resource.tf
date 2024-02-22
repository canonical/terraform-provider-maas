resource "maas_vm_host_machine" "kvm" {
  count   = 2
  vm_host = maas_vm_host.kvm.id
  cores   = 1
  memory  = 2048

  storage_disks {
    size_gigabytes = 15
  }
}
