data "maas_rack_controllers" "all" {}

# Use the first rack controller for DHCP configuration
resource "maas_subnet_ip_range" "example" {
  subnet   = data.maas_rack_controllers.all.controllers[0].subnets[0].id
  start_ip = "192.168.10.10"
  end_ip   = "192.168.10.50"
  type     = "dynamic"
}

resource "maas_vlan_dhcp" "dhcp" {
  fabric                  = data.maas_rack_controllers.all.controllers[0].subnets[0].fabric
  vlan                    = data.maas_rack_controllers.all.controllers[0].subnets[0].vlan
  ip_ranges               = [maas_subnet_ip_range.example.id]
  primary_rack_controller = data.maas_rack_controllers.all.controllers[0].id
}
