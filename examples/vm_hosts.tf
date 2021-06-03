resource "maas_vm_host" "kvm" {
  type = "virsh"
  power_address = "qemu+ssh://ubuntu@10.113.1.24/system"
  tags = [
    "pod-console-logging",
    "virtual",
    "kvm",
  ]
}

resource "maas_vm_host_machine" "kvm" {
  count = 2
  vm_host = maas_vm_host.kvm.id
  cores = 1
  memory = 2048
  storage = "disk1:15"
}

resource "maas_vm_host" "maas_machine" {
  type = "virsh"
  machine = "machine-01"
}

resource "maas_vm_host_machine" "maas_machine" {
  count = 3
  vm_host = maas_vm_host.maas_machine.id
}
