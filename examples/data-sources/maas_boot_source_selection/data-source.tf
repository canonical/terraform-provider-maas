data "maas_boot_source" "default" {}

data "maas_boot_source_selection" "default" {
  boot_source = maas_boot_source.default.boot_source

  os      = "ubuntu"
  release = "noble"
}
