data "maas_boot_source" "default" {}

data "maas_boot_source_selection" "default" {
  boot_source = data.maas_boot_source.default.id

  os      = "ubuntu"
  release = "noble"
}
