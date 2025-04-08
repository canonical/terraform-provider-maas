resource "maas_fabric" "test" {
  name = "tf-fabric"
}

data "maas_rack_controller" "test" {
  hostname = "maas-dev"
}

data "maas_vlan" "test" {
  vlan   = 0 # the default untagged vlan on all new fabrics.
  fabric = maas_fabric.test.id
}

resource "maas_subnet" "test" {
  cidr   = "10.100.100.0/24"
  fabric = maas_fabric.test.id
  vlan   = data.maas_vlan.test.id
}

resource "maas_subnet_ip_range" "test" {
  subnet   = maas_subnet.test.id
  start_ip = "10.100.100.2"
  end_ip   = "10.100.100.5"
  type     = "dynamic"
}

# VLAN 0 on test fabric with DHCP enabled, provided by primary rack controller.
resource "maas_vlan_dhcp" "test" {
  fabric                  = maas_fabric.test.id
  vlan                    = data.maas_vlan.test.vlan
  primary_rack_controller = data.maas_rack_controller.test.id
  ip_ranges               = [maas_subnet_ip_range.test.id]
}

resource "maas_fabric" "dummy" {
  name = "tf-fabric-dummy"
}

data "maas_vlan" "dummy" {
  vlan   = 0
  fabric = maas_fabric.dummy.id
}

resource "maas_subnet" "dummy" {
  cidr   = "10.100.102.0/24"
  fabric = maas_fabric.dummy.id
  vlan   = data.maas_vlan.dummy.vlan
}

resource "maas_subnet_ip_range" "dummy" {
  subnet   = maas_subnet.dummy.id
  start_ip = "10.100.102.1"
  end_ip   = "10.100.102.5"
  type     = "dynamic"
}

# VLAN 0 on dummy fabric with DHCP provided by relay VLAN on test fabric.
resource "maas_vlan_dhcp" "test_2" {
  fabric     = maas_fabric.dummy.id
  vlan       = data.maas_vlan.dummy.vlan
  ip_ranges  = [maas_subnet_ip_range.dummy.id]
  relay_vlan = data.maas_vlan.test.id
}
