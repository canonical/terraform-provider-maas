terraform {
  required_providers {
    maas = {
      source = "registry.terraform.io/ionutbalutoiu/maas"
    }
  }
}

provider "maas" {
  api_key = "<API_KEY>"
  api_url = "http://<MAAS_ADDRESS>:5240/MAAS"
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

resource "maas_instance" "kvm" {
  count = 2
  allocate_hostname = maas_pod_machine.kvm[count.index].hostname
  deploy_distro_series = "bionic"
}

output "maas_instance_kvm" {
  value = maas_instance.kvm
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

output "maas_machine_virsh_vm1" {
  value = maas_machine.virsh_vm1
}
output "maas_machine_virsh_vm2" {
  value = maas_machine.virsh_vm2
}
