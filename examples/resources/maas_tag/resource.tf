resource "maas_tag" "kvm" {
  name = "kvm"
  machines = [
    maas_machine.virsh_vm1.id,
    maas_machine.virsh_vm2.id,
    maas_vm_host_machine.kvm[0].id,
    maas_vm_host_machine.kvm[1].id,
  ]
}
