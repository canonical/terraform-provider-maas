resource "maas_zone" "test_zone" {
  description = "A description of the test zone"
  name        = "test-zone"
}

data "maas_zone" "test_zone" {
  name = maas_zone.test_zone.name
}
