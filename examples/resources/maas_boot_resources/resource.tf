resource "maas_boot_source" "test_boot_source" {
  url = "http://images.maas.io/ephemeral-v3/candidate/"
}

resource "maas_boot_source_selection" "jammy_test" {
  boot_source = maas_boot_source.test_boot_source.id

  os      = "ubuntu"
  release = "jammy"
}

resource "maas_boot_source_selection" "noble_test" {
  boot_source = maas_boot_source.test_boot_source.id

  os      = "ubuntu"
  release = "noble"
}

resource "maas_boot_resources" "test" {
  boot_source = maas_boot_source.test_boot_source.id

  boot_source_selections = [
    maas_boot_source_selection.jammy_test.id,
    maas_boot_source_selection.noble_test.id,
  ]
}
