
# Resource: maas_network_interface_link

Provides a resource to manage network configuration on a network interface.

## Example Usage

```terraform
resource "maas_network_interface_link" "virsh_vm1_nic1" {
  machine = maas_machine.virsh_vm1.id
  network_interface = maas_network_interface_physical.virsh_vm1_nic1.id
  subnet = data.maas_subnet.pxe.id
  mode = "STATIC"
  ip_address = "10.121.10.29"
  default_gateway = true
}
```

## Argument Reference

The following arguments are supported:

* `machine` - (Required) The identifier (system ID, hostname, or FQDN) of the machine with the network interface.
* `network_interface` - (Required) The identifier (MAC address, name, or ID) of the network interface.
* `subnet` - (Required) The identifier (CIDR or ID) of the subnet to be connected.
* `mode` - (Optional) Connection mode to subnet. It defaults to `AUTO`. Valid options are:
  * `AUTO` - Random static IP address from the subnet.
  * `DHCP` - IP address from the DHCP on the given subnet.
  * `STATIC` - Use `ip_address` as static IP address.
  * `LINK_UP` - Bring the interface up only on the given subnet. No IP address will be assigned.
* `default_gateway` - (Optional) Boolean value. When enabled, it sets the subnet gateway IP address as the default gateway for the machine the interface belongs to. This option can only be used with the `AUTO` and `STATIC` modes. Defaults to `false`.
* `ip_address` - (Optional) Valid IP address (from the given subnet) to be configured on the network interface. Only used when `mode` is set to `STATIC`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The network interface link ID.

## Import

This resource doesn't support the import operation.
