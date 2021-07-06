
# Resource: maas_machine

Provides a resource to manage MAAS machines.

## Example Usage

```terraform
resource "maas_machine" "virsh" {
  power_type = "virsh"
  power_parameters = {
    power_address = "qemu+ssh://ubuntu@10.113.1.10/system"
    power_id = "test-machine"
  }
  pxe_mac_address = "52:54:00:f9:11:e4"
}
```

## Argument Reference

The following arguments are supported:

* `power_type` - (Required) A power management type (e.g. `ipmi`).
* `power_parameters` - (Required) A map with the parameters specific to the `power_type`. See [Power types](https://maas.io/docs/api#power-types) section for a list of the available power parameters for each power type.
* `pxe_mac_address` - (Required) The MAC address of the machine's PXE boot NIC.
* `architecture` - (Optional) The architecture type of the machine. Defaults to `amd64/generic`.
* `min_hwe_kernel` - (Optional) The minimum kernel version allowed to run on this machine. Only used when deploying Ubuntu. This is computed if it's not set.
* `hostname` - (Optional) The machine hostname. This is computed if it's not set.
* `domain` - (Optional) The domain of the machine. This is computed if it's not set.
* `zone` - (Optional) The zone of the machine. This is computed if it's not set.
* `pool` - (Optional) The resource pool of the machine. This is computed if it's not set.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The MAAS machine system ID.

## Import

MAAS machines can be imported using one of the attributes: system ID, hostname, or FQDN. e.g.

```shell
terraform import maas_machine.virsh_vm1 vm1.maas
```
