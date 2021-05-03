data "maas_fabric" "default" {
  name = "maas"
}

data "maas_vlan" "default" {
  fabric_id = data.maas_fabric.default.id
  vid = 0
}

data "maas_vlan" "vlan10" {
  fabric_id = data.maas_fabric.default.id
  vid = 10
}

data "maas_subnet" "pxe" {
  cidr = "10.121.0.0/16"
  vlan_id = data.maas_vlan.default.id
}

data "maas_subnet" "vlan10" {
  cidr = "10.10.0.0/16"
  vlan_id = data.maas_vlan.vlan10.id
}

resource "maas_machine" "virsh_vm1" {
  power_type = "virsh"
  power_parameters = {
    power_address = "qemu+ssh://ibalutoiu@10.121.0.10/system"
    power_id = "test-vm1"
  }
  pxe_mac_address = "52:54:00:89:f5:3e"
}

resource "maas_network_interface_physical" "virsh_vm1_nic1" {
  machine_id = maas_machine.virsh_vm1.id
  mac_address = "52:54:00:89:f5:3e"
  name = "eth0"
  vlan = data.maas_vlan.default.id
  tags = [
    "nic1-tag1",
    "nic1-tag2",
    "nic1-tag3",
  ]
}

resource "maas_network_interface_physical" "virsh_vm1_nic2" {
  machine_id = maas_machine.virsh_vm1.id
  mac_address = "52:54:00:f5:89:ae"
  name = "eth1"
  vlan = data.maas_vlan.vlan10.id
  tags = [
    "nic2-tag1",
    "nic2-tag2",
    "nic2-tag3",
  ]
}

resource "maas_network_interface_physical" "virsh_vm1_nic3" {
  machine_id = maas_machine.virsh_vm1.id
  mac_address = "52:54:00:0e:92:79"
  name = "eth2"
  vlan = data.maas_vlan.default.id
  tags = [
    "nic3-tag1",
    "nic3-tag2",
    "nic3-tag3",
  ]
}

resource "maas_machine" "virsh_vm2" {
  power_type = "virsh"
  power_parameters = {
    power_address = "qemu+ssh://ibalutoiu@10.121.0.10/system"
    power_id = "test-vm2"
  }
  pxe_mac_address = "52:54:00:7c:f7:77"
}

resource "maas_network_interface_physical" "virsh_vm2_nic1" {
  machine_id = maas_machine.virsh_vm2.id
  mac_address = "52:54:00:7c:f7:77"
  name = "eno0"
  vlan = data.maas_vlan.default.id
  tags = [
    "nic1-tag1",
    "nic1-tag2",
    "nic1-tag3",
  ]
}

resource "maas_network_interface_physical" "virsh_vm2_nic2" {
  machine_id = maas_machine.virsh_vm2.id
  mac_address = "52:54:00:82:5c:c1"
  name = "eno1"
  vlan = data.maas_vlan.default.id
  tags = [
    "nic2-tag1",
    "nic2-tag2",
    "nic2-tag3",
  ]
}

resource "maas_network_interface_physical" "virsh_vm2_nic3" {
  machine_id = maas_machine.virsh_vm2.id
  mac_address = "52:54:00:bb:6e:9f"
  name = "eno2"
  vlan = data.maas_vlan.vlan10.id
  tags = [
    "nic3-tag1",
    "nic3-tag2",
    "nic3-tag3",
  ]
}

resource "maas_pod" "kvm" {
  type = "virsh"
  power_address = "qemu+ssh://ubuntu@10.113.1.10/system"
  tags = [
    "pod-console-logging",
    "virtual",
    "kvm",
  ]
}

resource "maas_pod_machine" "kvm" {
  count = 2
  pod = maas_pod.kvm.id
  cores = 1
  memory = 2048
  storage = "disk1:15"
}

resource "maas_tag" "kvm" {
  name = "kvm"
  machine_ids = [
    maas_pod_machine.kvm[0].id,
    maas_pod_machine.kvm[1].id,
    maas_machine.virsh_vm1.id,
    maas_machine.virsh_vm2.id,
  ]
}

resource "maas_tag" "virtual" {
  name = "virtual"
  machine_ids = [
    maas_pod_machine.kvm[0].id,
    maas_pod_machine.kvm[1].id,
    maas_machine.virsh_vm1.id,
    maas_machine.virsh_vm2.id,
  ]
}

resource "maas_tag" "ubuntu" {
  name = "ubuntu"
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
    maas_tag.ubuntu.name,
  ]
  deploy_distro_series = "focal"
  depends_on = [
    maas_network_interface_physical.virsh_vm1_nic1,
    maas_network_interface_physical.virsh_vm1_nic2,
    maas_network_interface_physical.virsh_vm1_nic3,
    maas_network_interface_physical.virsh_vm2_nic1,
    maas_network_interface_physical.virsh_vm2_nic2,
    maas_network_interface_physical.virsh_vm2_nic3,
  ]
}

output "maas_network_interface_physical-virsh_vm1_nic1" { value = maas_network_interface_physical.virsh_vm1_nic1 }
output "maas_network_interface_physical-virsh_vm1_nic2" { value = maas_network_interface_physical.virsh_vm1_nic2 }
output "maas_network_interface_physical-virsh_vm1_nic3" { value = maas_network_interface_physical.virsh_vm1_nic3 }
output "maas_network_interface_physical-virsh_vm2_nic1" { value = maas_network_interface_physical.virsh_vm2_nic1 }
output "maas_network_interface_physical-virsh_vm2_nic2" { value = maas_network_interface_physical.virsh_vm2_nic2 }
output "maas_network_interface_physical-virsh_vm2_nic3" { value = maas_network_interface_physical.virsh_vm2_nic3 }
output "maas_machine_virsh_vm1-hostname" { value = maas_machine.virsh_vm1.hostname }
output "maas_machine_virsh_vm2-hostname" { value = maas_machine.virsh_vm2.hostname }
output "maas_instance_kvm" { value = maas_instance.kvm }
