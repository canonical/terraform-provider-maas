resource "maas_block_device_tag" "test" {
  machine         = "abc123"
  block_device_id = 12
  tags = [
    "tag1",
    "tag2",
  ]
}

resource "maas_block_device_tag" "test2" {
  machine         = "amazed-kiwi"
  block_device_id = 13
  tags = [
    "tag3",
    "tag4",
  ]
}

