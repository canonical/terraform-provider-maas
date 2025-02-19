data "maas_boot_source" "default" {}

data "maas_boot_source_selection" "default" {
  boot_source = maas_boot_source.default.id

  os      = "ubuntu"
  release = "noble"
}
