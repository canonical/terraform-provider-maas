data "maas_subnet" "pxe" {
  cidr = "10.99.0.0/16"
}

data "maas_subnet" "vid10" {
  cidr = "10.10.0.0/16"
}
