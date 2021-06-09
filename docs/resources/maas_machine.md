
# `maas_machine`

Creates a new MAAS machine.

Example:

```hcl
resource "maas_machine" "virsh" {
  power_type = "virsh"
  power_parameters = {
    power_address = "qemu+ssh://ubuntu@10.113.1.10/system"
    power_id = "test-machine"
  }
  pxe_mac_address = "52:54:00:f9:11:e4"
}
```

Parameters:

| Name | Type | Required | Description
| ---- | ---- | -------- | -----------
| `power_type` | `string` | `true` | A power management type (e.g. `virsh`, `ipmi`).
| `power_parameters` | `map[string]string` | `true` | The parameter(s) for the `power_type`. Note that this is dynamic as the available parameters depend on the selected value of the Machine's `power_type`. See [Power types](https://maas.io/docs/api#power-types) section for a list of the available power parameters for each power type.
| `pxe_mac_address` | `string` | `true` | The MAC address of the machine's PXE boot NIC.
| `architecture` | `string` | `false` | A string containing the architecture type of the machine.
| `min_hwe_kernel` | `string` | `false` | A string containing the minimum kernel version allowed to be ran on this machine.
| `hostname` | `string` | `false` | A hostname. If not given, one will be generated.
| `domain` | `string` | `false` | The domain of the machine. If not given, the default domain is used.
| `zone` | `string` | `false` | Name of a valid physical zone in which to place this machine.
| `pool` | `string` | `false` | The resource pool to which the machine should belong.
