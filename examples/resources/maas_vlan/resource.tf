resource "maas_vlan" "tf_vlan" {
  fabric = maas_fabric.tf_fabric.id
  vid    = 14
  name   = "tf-vlan14"
  space  = maas_space.tf_space.name
}
