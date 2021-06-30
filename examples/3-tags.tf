resource "maas_tag" "kvm" {
  name = "kvm"
  machine_ids = [
    maas_machine.virsh_vm1.id,
    maas_machine.virsh_vm2.id,
    maas_vm_host_machine.kvm[0].id,
    maas_vm_host_machine.kvm[1].id,
  ]
}

resource "maas_tag" "virtual" {
  name = "virtual"
  machine_ids = [
    maas_machine.virsh_vm1.id,
    maas_machine.virsh_vm2.id,
    maas_vm_host_machine.kvm[0].id,
    maas_vm_host_machine.kvm[1].id,
  ]
}

resource "maas_tag" "ubuntu" {
  name = "ubuntu"
  machine_ids = [
    maas_machine.virsh_vm1.id,
    maas_machine.virsh_vm2.id,
    maas_vm_host_machine.kvm[0].id,
    maas_vm_host_machine.kvm[1].id,
  ]
}
