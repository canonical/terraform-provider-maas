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
  allocate_min_cpu_count = 1
  allocate_min_memory = 2048
  allocate_zone = "default"
  allocate_pool = "default"
  allocate_tags = [
    "virtual",
    "kvm",
  ]
  deploy_distro_series = "focal"
  deploy_hwe_kernel = "focal (ga-20.04)"  # Only used when deploying Ubuntu.
  deploy_user_data = "${file("${path.module}/user-data.txt")}"
}

output "maas_instance_two_machines" {
  value = maas_instance.two_machines
}
