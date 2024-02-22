resource "maas_subnet_ip_range" "dynamic_ip_range" {
  subnet   = maas_subnet.tf_subnet_2.id
  type     = "dynamic"
  start_ip = "10.77.77.2"
  end_ip   = "10.77.77.60"
}

resource "maas_subnet_ip_range" "reserved_ip_range" {
  subnet   = maas_subnet.tf_subnet_2.id
  type     = "reserved"
  start_ip = "10.77.77.200"
  end_ip   = "10.77.77.254"
  comment  = "Reserved for Static IPs"
}
