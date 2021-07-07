
# Resource: maas_vlan

Provides a resource to manage MAAS network VLANs.

## Example Usage

```terraform
resource "maas_vlan" "tf_vlan" {
  fabric = maas_fabric.tf_fabric.id
  vid = 14
  name = "tf-vlan14"
  space = maas_space.tf_space.name
}
```

## Argument Reference

The following arguments are supported:

* `fabric` - (Required) The identifier (name or ID) of the fabric for the new VLAN.
* `vid` - (Required) The traffic segregation ID for the new VLAN.
* `mtu` - (Optional) The MTU to use on the new VLAN. This argument is computed if it's not set.
* `dhcp_on` - (Optional) Boolean value. Whether or not DHCP should be managed on the new VLAN. This argument is computed if it's not set.
* `name` - (Optional) The name of the new VLAN. This argument is computed if it's not set.
* `space` - (Optional) The space of the new VLAN. Passing in an empty string (or the string `undefined`) will cause the VLAN to be placed in the `undefined` space. This argument is computed if it's not set.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The VLAN MAAS resource ID.

## Import

Existing MAAS VLANs can be imported using the fabric identifier (ID or name) and the VLAN identifier (ID or traffic segregation ID). e.g.

```shell
terraform import maas_vlan.tf_vlan tf-fabric:14
```
