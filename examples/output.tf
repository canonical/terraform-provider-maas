# Network
output "maas_fabric" { value = data.maas_fabric.default }
output "maas_vlan_untagged" { value = data.maas_vlan.default }
output "maas_vlan_10" { value = data.maas_vlan.vid10 }
output "maas_subnet_pxe" { value = data.maas_subnet.pxe }
output "maas_subnet_vlan_10" { value = data.maas_subnet.vid10 }
output "maas_space_tf_space" { value = maas_space.tf_space }
output "maas_fabric_tf_fabric" { value = maas_fabric.tf_fabric }
output "maas_vlan_tf_vlan" { value = maas_vlan.tf_vlan }
output "maas_subnet_tf_subnet" { value = maas_subnet.tf_subnet }
output "maas_subnet_tf_subnet_2" { value = maas_subnet.tf_subnet_2 }
output "maas_subnet_ip_range_dynamic" { value = maas_subnet_ip_range.dynamic_ip_range }
output "maas_subnet_ip_range_reserved" { value = maas_subnet_ip_range.reserved_ip_range }
output "maas_dns_domain_cloudbase" { value = maas_dns_domain.cloudbase }
output "maas_dns_record_test_a" { value = maas_dns_record.test_a }
output "maas_dns_record_test_aaaa" { value = maas_dns_record.test_aaaa }
output "maas_dns_record_test_txt" { value = maas_dns_record.test_txt }

# Machines
output "maas_machine_1" { value = maas_machine.virsh_vm1.hostname }
output "maas_machine_2" { value = maas_machine.virsh_vm2.hostname }

# VM Hosts
output "maas_vm_host_kvm" { value = maas_vm_host.kvm.name }
output "maas_vm_host_kvm_1" { value = maas_vm_host_machine.kvm[0] }
output "maas_vm_host_kvm_2" { value = maas_vm_host_machine.kvm[1] }
output "maas_vm_host_maas_machine" { value = maas_vm_host.maas_machine.name }
output "maas_vm_host_maas_machine_1" { value = maas_vm_host_machine.maas_machine_1 }
output "maas_vm_host_maas_machine_2" { value = maas_vm_host_machine.maas_machine_2 }

# Tags
output "maas_tag_kvm" { value = maas_tag.kvm }
output "maas_tag_virtual" { value = maas_tag.virtual }
output "maas_tag_ubuntu" { value = maas_tag.ubuntu }

# Instances
output "maas_instance_1" { value = maas_instance.kvm[0] }
output "maas_instance_2" { value = maas_instance.kvm[1] }
output "maas_instance_virsh_vm4" { value = maas_instance.virsh_vm4 }
