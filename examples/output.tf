# Network
output "maas_fabric"         { value = data.maas_fabric.default }
output "maas_vlan_untagged"  { value = data.maas_vlan.default }
output "maas_vlan_10"        { value = data.maas_vlan.vid10 }
output "maas_subnet_pxe"     { value = data.maas_subnet.pxe }
output "maas_subnet_vlan_10" { value = data.maas_subnet.vid10 }

# Machines
output "maas_machine_1" { value = maas_machine.virsh_vm1.hostname }
output "maas_machine_2" { value = maas_machine.virsh_vm2.hostname }

# VM Hosts
output "maas_vm_host_kvm"             { value = maas_vm_host.kvm.name }
output "maas_vm_host_kvm_1"           { value = maas_vm_host_machine.kvm[0] }
output "maas_vm_host_kvm_2"           { value = maas_vm_host_machine.kvm[1] }
output "maas_vm_host_maas_machine"    { value = maas_vm_host.maas_machine.name }
output "maas_vm_host_maas_machine_1"  { value = maas_vm_host_machine.maas_machine[0] }
output "maas_vm_host_maas_machine_2"  { value = maas_vm_host_machine.maas_machine[1] }
output "maas_vm_host_maas_machine_3"  { value = maas_vm_host_machine.maas_machine[2] }

# Tags
output "maas_tag_kvm"     { value = maas_tag.kvm }
output "maas_tag_virtual" { value = maas_tag.virtual }
output "maas_tag_ubuntu"  { value = maas_tag.ubuntu }

# Instances
output "maas_instance_1" { value = maas_instance.kvm[0] }
output "maas_instance_2" { value = maas_instance.kvm[1] }
