resource "maas_network_interface_tag" "test" {
  machine      = "abc123"
  interface_id = 12
  tags = [
    "tag1",
    "tag2",
  ]
}

resource "maas_network_interface_tag" "test2" {
  device       = "cheerful-owl"
  interface_id = 13
  tags = [
    "tag3",
    "tag4",
  ]
}

resource "maas_network_interface_tag" "test3" {
  device       = "def456"
  interface_id = 14
  tags = [
    "tag3",
    "tag4",
  ]
}

