resource "maas_boot_source" "test_boot_source" {
  url = "http://images.maas.io/ephemeral-v3/candidate/"
}

resource "maas_boot_source_selection" "test" {
  boot_source = maas_boot_source.test_boot_source.id

  os      = "ubuntu"
  release = "jammy"
}
