# `maas_network_interface_link`

Configures a machine network interface with an IP address from a given subnet.

Example:

```hcl
resource "maas_network_interface_link" "virsh_vm1_nic1" {
  machine_id = maas_machine.virsh_vm1.id
  network_interface_id = maas_network_interface_physical.virsh_vm1_nic1.id
  subnet_id = data.maas_subnet.pxe.id
  mode = "STATIC"
  ip_address = "10.121.10.29"
  default_gateway = true
}
```

Parameters:

| Name | Type | Required | Description
| ---- | ---- | -------- | -----------
| `machine_id` | `string` | `true` | Machine system id.
| `network_interface_id` | `int` | `true` | Network interface id.
| `subnet_id` | `int` | `true` | Subnet id.
| `mode` | `string` | `false` | Connection mode to subnet. It defaults to `AUTO`. Valid options are: `AUTO` (random static IP address from the subnet), `DHCP` (DHCP on the given subnet), `STATIC` (use `ip_address` as static IP address).
| `ip_address` | `string` | `false` | IP address for the interface in the given subnet. Only used when `mode` is `STATIC`.
| `default_gateway` | `bool` | `false` | When enabled, it sets the subnet gateway IP address as the default gateway for the machine this interface belongs to. This option can only be used with the `AUTO` and `STATIC` modes.
