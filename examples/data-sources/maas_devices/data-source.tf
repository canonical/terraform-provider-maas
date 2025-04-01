resource "maas_device" "test_device" {
  description = "Test description"
  domain      = "maas"
  hostname    = "test-device"
  zone        = "default"
  network_interfaces {
    mac_address = "12:23:45:67:89:de"
  }
}

data "maas_devices" "all" {}
