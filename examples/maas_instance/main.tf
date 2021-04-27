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

resource "maas_instance" "two_machines" {
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

output "maas_instance_two_machines" {
  value = maas_instance.two_machines
}
