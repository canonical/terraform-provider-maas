resource "maas_reserved_ip" "my_server_ip" {
  ip          = "10.0.0.50"
  mac_address = "aa:bb:cc:dd:ee:ff"
  subnet      = maas_subnet.example.id
  comment     = "Server static lease"
}

resource "maas_reserved_ip" "another_ip" {
  ip          = "10.0.0.51"
  mac_address = "11:22:33:44:55:66"
}
