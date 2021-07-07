
# Resource: maas_vm_host

Provides a resource to manage MAAS VM hosts.

## Example Usage

### Using pre-deployed VM host

```terraform
resource "maas_vm_host" "kvm" {
  type = "virsh"
  power_address = "qemu+ssh://ubuntu@10.113.1.10/system"
}
```

### Deploy a new VM host from a ready MAAS machine

```terraform
resource "maas_vm_host" "maas_machine" {
  type = "virsh"
  machine = "machine-05"
}
```

## Argument Reference

The following arguments are supported:

* `type` - (Required) The VM host type. Supported values are: `lxd`, `virsh`.
* `machine` - (Optional) The identifier (hostname, FQDN or system ID) of a registered ready MAAS machine. This is going to be deployed and registered as a new VM host. This argument conflicts with: `power_address`, `power_user`, `power_pass`.
* `power_address` - (Optional) Address that gives MAAS access to the VM host power control. For example: `qemu+ssh://172.16.99.2/system`. The address given here must reachable by the MAAS server. It can't be set if `machine` argument is used.
* `power_user` - (Optional) User name to use for power control of the VM host. Cannot be set if `machine` parameter is used.
* `power_pass` - (Optional) User password to use for power control of the VM host. Cannot be set if `machine` parameter is used.
* `name` - (Optional) The new VM host name. This is computed if it's not set.
* `zone` - (Optional) The new VM host zone name. This is computed if it's not set.
* `pool` - (Optional) The new VM host pool name. This is computed if it's not set.
* `tags` - (Optional) A set of tag names to assign to the new VM host. This is computed if it's not set.
* `cpu_over_commit_ratio` - (Optional) The new VM host CPU overcommit ratio. This is computed if it's not set.
* `memory_over_commit_ratio` - (Optional) The new VM host RAM memory overcommit ratio. This is computed if it's not set.
* `default_macvlan_mode` - (Optional) The new VM host default macvlan mode. Supported values are: `bridge`, `passthru`, `private`, `vepa`. This is computed if it's not set.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The VM host ID.
* `resources_cores_total` - The VM host total number of CPU cores.
* `resources_memory_total` - The VM host total RAM memory (in MB).
* `resources_local_storage_total` - The VM host total local storage (in bytes).

## Import

VM hosts can be imported using the ID or the name. e.g.

```shell
terraform import maas_vm_host.kvm vm-host-01
```
