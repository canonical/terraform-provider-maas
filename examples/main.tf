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
        "virtual"
    ]
}

output "ibalutoiu_maas_instances" {
  value = maas_instance.ibalutoiu
}
