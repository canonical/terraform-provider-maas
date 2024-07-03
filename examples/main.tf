terraform {
  required_providers {
    maas = {
      source = "registry.terraform.io/canonical/maas"
    }
  }
}

provider "maas" {}
