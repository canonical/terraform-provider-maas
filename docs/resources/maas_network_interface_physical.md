# `maas_network_interface_physical`

Configures a physical network interface from an existing MAAS machine.

Example:

```hcl
resource "maas_network_interface_physical" "virsh_vm1_nic1" {
  machine_id = maas_machine.virsh_vm1.id
  mac_address = "52:54:00:89:f5:3e"
  name = "eth0"
  vlan = data.maas_vlan.default.id
  tags = [
    "nic1-tag1",
    "nic1-tag2",
    "nic1-tag3",
  ]
}
```

Parameters:

| Name | Type | Required | Description
| ---- | ---- | -------- | -----------
| `machine_id` | `string` | `true` | Machine system id.
| `mac_address` | `string` | `true` | The physical networking interface MAC address.
| `name` | `string` | `false` | The physical networking interface name.
| `tags` | `[]string` | `false` | Tags for the interface.
| `vlan` | `string` | `false` | VLAN the interface is connected to. Defaults to `untagged`.
| `mtu` | `int` | `false` | Maximum transmission unit. Defaults to `1500`.
| `accept_ra` | `bool` | `false` | Accept router advertisements (IPv6 only).
| `autoconf` | `bool` | `false` | Perform stateless autoconfiguration (IPv6 only).
