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

output "maas_pod_kvm-type"                              {value = maas_pod.kvm.type}
output "maas_pod_kvm-power_address"                     {value = maas_pod.kvm.power_address}
output "maas_pod_kvm-power_user"                        {value = maas_pod.kvm.power_user}
output "maas_pod_kvm-name"                              {value = maas_pod.kvm.name}
output "maas_pod_kvm-zone"                              {value = maas_pod.kvm.zone}
output "maas_pod_kvm-pool"                              {value = maas_pod.kvm.pool}
output "maas_pod_kvm-tags"                              {value = maas_pod.kvm.tags}
output "maas_pod_kvm-cpu_over_commit_ratio"             {value = maas_pod.kvm.cpu_over_commit_ratio}
output "maas_pod_kvm-memory_over_commit_ratio"          {value = maas_pod.kvm.memory_over_commit_ratio}
output "maas_pod_kvm-default_macvlan_mode"              {value = maas_pod.kvm.default_macvlan_mode}
output "maas_pod_kvm-resources_cores_available"         {value = maas_pod.kvm.resources_cores_available}
output "maas_pod_kvm-resources_cores_total"             {value = maas_pod.kvm.resources_cores_total}
output "maas_pod_kvm-resources_memory_available"        {value = maas_pod.kvm.resources_memory_available}
output "maas_pod_kvm-resources_memory_total"            {value = maas_pod.kvm.resources_memory_total}
output "maas_pod_kvm-resources_local_storage_available" {value = maas_pod.kvm.resources_local_storage_available}
output "maas_pod_kvm-resources_local_storage_total"     {value = maas_pod.kvm.resources_local_storage_total}

resource "maas_pod_machine" "kvm" {
  count = 3
  pod = maas_pod.kvm.id
  cores = 1
  memory = 2048
  storage = "disk1:32,disk2:20"
  domain = "maas"
  zone = "default"
  pool = "default"
}

output "maas_pod_machine_kvm" {
  value = maas_pod_machine.kvm
}
