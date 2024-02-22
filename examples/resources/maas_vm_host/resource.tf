resource "maas_vm_host" "kvm" {
  type          = "virsh"
  power_address = "qemu+ssh://ubuntu@10.113.1.24/system"
  tags = [
    "pod-console-logging",
    "virtual",
    "kvm",
  ]
}
