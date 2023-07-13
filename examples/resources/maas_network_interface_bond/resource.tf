resource "maas_network_interface_bond" "test" {
  machine        = maas_machine.example.id
  name           = "bond0"
  accept_ra      = false
  bond_lacp_rate = "slow"
  bond_mode      = "802.3ad"
  mtu            = 9000
  parents        = ["eth1", "eth2"]
}
