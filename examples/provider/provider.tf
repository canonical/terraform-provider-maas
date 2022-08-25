terraform {
  required_providers {
    maas = {
      source  = "maas/maas"
      version = "~>1.0"
    }
  }
}

provider "maas" {
  api_version = "2.0"
  api_key = "<YOUR API KEY>"
  api_url = "http://127.0.0.1:5240/MAAS"
}

