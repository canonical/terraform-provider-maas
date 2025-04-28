resource "maas_boot_source" "test_boot_source" {
  url = "http://images.maas.io/ephemeral-v3/candidate/"
}

resource "maas_boot_source_selection" "amd64" {
  boot_source = maas_boot_source.test_boot_source.id

  os      = "ubuntu"
  release = "jammy"
  arches  = ["amd64"]
}

# Only Hardware Enablement kernel (HWE) image
resource "maas_boot_source_selection" "amd64_hwe" {
  boot_source = maas_boot_source.test_boot_source.id

  os        = "ubuntu"
  release   = "jammy"
  arches    = ["amd64"]
  subarches = ["hwe-22.04"]
}
