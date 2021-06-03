
# `maas_vm_host`

Creates a new MAAS VM host.

Example:

```hcl
resource "maas_vm_host" "kvm" {
  type = "virsh"
  power_address = "qemu+ssh://ubuntu@10.113.1.10/system"
  name = "kvm-host-01"
}
```

Parameters:

| Name | Type | Required | Description
| ---- | ---- | -------- | -----------
| `type` | `string` | `true` | The type of VM host to create: `lxd` or `virsh`.
| `machine` | `string` | `false` | The identifier (`hostname`, `fqdn` or `system_id`) of a registered `Ready` MAAS machine. This is going to be deployed and registered as a new VM host.
| `power_address` | `string` | `false` | Address that gives MAAS access to the VM host power control. For example: `qemu+ssh://172.16.99.2/system`. Cannot be set if `machine` parameter is used.
| `power_user` | `string` | `false` | Username to use for power control of the VM host. Cannot be set if `machine` parameter is used.
| `power_pass` | `string` | `false` | Password to use for power control of the VM host. Cannot be set if `machine` parameter is used.
| `name` | `string` | `false` | The new VM host name.
| `zone` | `string` | `false` | The new VM host zone.
| `pool` | `string` | `false` | The name of the resource pool the new VM host will belong to. Machines composed from this VM host will be assigned to this resource pool by default.
| `tags` | `[]string` | `false` | A list of tags to assign to the new VM host.
| `cpu_over_commit_ratio` | `float` | `false` | CPU overcommit ratio.
| `memory_over_commit_ratio` | `float` | `false` | RAM memory overcommit ratio.
| `default_macvlan_mode` | `string` | `false` |  Default macvlan mode for VM hosts that use it: `bridge`, `passthru`, `private`, `vepa`.
