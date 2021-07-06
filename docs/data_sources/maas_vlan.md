# Data Source: maas_vlan

Provides details about an existing MAAS VLAN.

## Example Usage

```terraform
data "maas_vlan" "vid10" {
  fabric = data.maas_fabric.default.id
  vlan = 10
}
```

## Argument Reference

The following arguments are supported:

* `fabric` - (Required) The fabric identifier (ID or name) for the VLAN.
* `vlan` - (Required) The VLAN identifier (ID or traffic segregation ID).

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `mtu` - The MTU used on the VLAN.
* `dhcp_on` - Boolean value indicating if DHCP is enabled on the VLAN.
* `name` - The VLAN name.
* `space` - The VLAN space.
