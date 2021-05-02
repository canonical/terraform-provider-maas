data "maas_subnet" "pxe" {
  cidr = "10.121.0.0/16"
}

data "maas_subnet" "vlan10" {
  cidr = "10.10.0.0/16"
  vid = 10
  fabric = "maas"
}

resource "maas_machine" "virsh_vm1" {
  power_type = "virsh"
  power_parameters = {
    power_address = "qemu+ssh://ibalutoiu@10.121.0.10/system"
    power_id = "test-vm1"
  }
  pxe_mac_address = "52:54:00:89:f5:3e"
}

resource "maas_machine" "virsh_vm2" {
  power_type = "virsh"
  power_parameters = {
    power_address = "qemu+ssh://ibalutoiu@10.121.0.10/system"
    power_id = "test-vm2"
  }
  pxe_mac_address = "52:54:00:7c:f7:77"
}

resource "maas_pod" "kvm" {
  type = "virsh"
  power_address = "qemu+ssh://ubuntu@10.113.1.10/system"
}

resource "maas_pod_machine" "kvm" {
  count = 2
  pod = maas_pod.kvm.id
  cores = 1
  memory = 2048
  storage = "disk1:15"
}

resource "maas_tag" "kvm" {
  name = "tf-kvm"
  machine_ids = [
    maas_pod_machine.kvm[0].id,
    maas_pod_machine.kvm[1].id,
    maas_machine.virsh_vm1.id,
    maas_machine.virsh_vm2.id,
  ]
}

resource "maas_tag" "virtual" {
  name = "tf-virtual"
  machine_ids = [
    maas_pod_machine.kvm[0].id,
    maas_pod_machine.kvm[1].id,
    maas_machine.virsh_vm1.id,
    maas_machine.virsh_vm2.id,
  ]
}

resource "maas_instance" "kvm" {
  count = 2
  allocate_hostname = maas_pod_machine.kvm[count.index].hostname
  allocate_min_cpu_count = 1
  allocate_min_memory = 2048
  allocate_tags = [
    maas_tag.virtual.name,
    maas_tag.kvm.name,
  ]
  deploy_distro_series = "focal"
}

output "maas_machine_virsh_vm1-hostname" { value = maas_machine.virsh_vm1.hostname }
output "maas_machine_virsh_vm2-hostname" { value = maas_machine.virsh_vm2.hostname }
output "maas_instance_kvm" { value = maas_instance.kvm }
