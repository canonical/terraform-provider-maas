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

resource "maas_instance" "ibalutoiu" {
  count = 2
  min_cpu_count = 1
  min_memory = 2048
  tags = [
      "virtual",
      "kvm",
  ]
  zone = "default"
  pool = "default"
  distro_series = "focal"
  hwe_kernel = "focal (ga-20.04)"  # Only used when deploying Ubuntu.
  user_data = "${file("${path.module}/user-data.txt")}"
}

output "ibalutoiu_maas_instances" {
  value = maas_instance.ibalutoiu
}
