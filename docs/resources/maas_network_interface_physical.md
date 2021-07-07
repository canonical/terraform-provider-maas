
# Resource: maas_network_interface_physical

Provides a resource to manage a physical network interface from an existing MAAS machine.

## Example Usage

```terraform
resource "maas_network_interface_physical" "virsh_vm1_nic1" {
  machine = maas_machine.virsh_vm1.id
  mac_address = "52:54:00:89:f5:3e"
  vlan = data.maas_vlan.default.id
  name = "eth0"
  mtu = 1450
  tags = [
    "nic1-tag1",
    "nic1-tag2",
    "nic1-tag3",
  ]
}
```

## Argument Reference

The following arguments are supported:

* `machine` - (Required) The identifier (system ID, hostname, or FQDN) of the machine with the physical network interface.
* `mac_address` - (Required) The physical network interface MAC address.
* `vlan` - (Optional) VLAN the physical network interface is connected to. Defaults to `untagged`.
* `name` - (Optional) The physical network interface name. This argument is computed if it's not set.
* `mtu` - (Optional) The MTU of the physical network interface. This argument is computed if it's not set.
* `tags` - (Optional) A set of tag names to be assigned to the physical network interface. This argument is computed if it's not set.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The physical network interface ID.

## Import

A physical network interface can be imported using the machine identifier (system ID, hostname, or FQDN) and its own identifier (MAC address, name, or ID). e.g.

```shell
terraform import maas_network_interface_physical.virsh_vm1 vm1:eth0
```
