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

resource "maas_vm_host" "kvm" {
  type = "virsh"
  power_address = "qemu+ssh://ubuntu@10.113.1.10/system"
  power_pass = "super-secure-password"
  name = "kvm-host-01"
  zone = "default"
  pool = "default"
  tags = [
    "pod-console-logging",
    "virtual",
    "kvm",
  ]
  cpu_over_commit_ratio = 2.5
  memory_over_commit_ratio = 1.5
  default_macvlan_mode = "bridge"
}

output "maas_vm_host_kvm-type"                              {value = maas_vm_host.kvm.type}
output "maas_vm_host_kvm-power_address"                     {value = maas_vm_host.kvm.power_address}
output "maas_vm_host_kvm-power_user"                        {value = maas_vm_host.kvm.power_user}
output "maas_vm_host_kvm-name"                              {value = maas_vm_host.kvm.name}
output "maas_vm_host_kvm-zone"                              {value = maas_vm_host.kvm.zone}
output "maas_vm_host_kvm-pool"                              {value = maas_vm_host.kvm.pool}
output "maas_vm_host_kvm-tags"                              {value = maas_vm_host.kvm.tags}
output "maas_vm_host_kvm-cpu_over_commit_ratio"             {value = maas_vm_host.kvm.cpu_over_commit_ratio}
output "maas_vm_host_kvm-memory_over_commit_ratio"          {value = maas_vm_host.kvm.memory_over_commit_ratio}
output "maas_vm_host_kvm-default_macvlan_mode"              {value = maas_vm_host.kvm.default_macvlan_mode}
output "maas_vm_host_kvm-resources_cores_available"         {value = maas_vm_host.kvm.resources_cores_available}
output "maas_vm_host_kvm-resources_cores_total"             {value = maas_vm_host.kvm.resources_cores_total}
output "maas_vm_host_kvm-resources_memory_available"        {value = maas_vm_host.kvm.resources_memory_available}
output "maas_vm_host_kvm-resources_memory_total"            {value = maas_vm_host.kvm.resources_memory_total}
output "maas_vm_host_kvm-resources_local_storage_available" {value = maas_vm_host.kvm.resources_local_storage_available}
output "maas_vm_host_kvm-resources_local_storage_total"     {value = maas_vm_host.kvm.resources_local_storage_total}

resource "maas_vm_host_machine" "kvm" {
  count = 3
  vm_host = maas_vm_host.kvm.id
  cores = 1
  memory = 2048
  storage = "disk1:32,disk2:20"
  domain = "maas"
  zone = "default"
  pool = "default"
}

output "maas_vm_host_machine_kvm" {
  value = maas_vm_host_machine.kvm
}
