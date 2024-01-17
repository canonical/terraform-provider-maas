resource "maas_resource_pool" "test_resource_pool" {
  description = "Test description"
  name        = "test-resource-pool"
}

data "maas_resource_pool" "test_resource_pool" {
  name = maas_resource_pool.test_resource_pool.name
}
