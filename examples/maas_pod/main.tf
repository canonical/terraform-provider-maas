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
    "virtual",
    "kvm",
  ]
  cpu_over_commit_ratio = 2.5
  memory_over_commit_ratio = 1.5
  default_macvlan_mode = "bridge"
}
