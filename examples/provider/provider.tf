terraform {
  required_providers {
    maas = {
      source  = "ionutbalutoiu/maas" // TODO: change to maas/maas after release
      version = "~>1.0"
    }
  }
}

provider "maas" {
  api_version = "2.0"
  api_key = "<YOUR API KEY>"
  api_url = "http://localhost:5240/MAAS"
}

