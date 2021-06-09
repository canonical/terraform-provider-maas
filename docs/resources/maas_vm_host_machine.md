# `maas_vm_host_machine`

Composes a new MAAS machine from an existing MAAS VM host.

Example:

```hcl
resource "maas_vm_host_machine" "kvm" {
  vm_host = maas_vm_host.kvm.id
  cores = 1
  memory = 2048
  storage = "disk1:32,disk2:20"
}
```

Parameters:

| Name | Type | Required | Description
| ---- | ---- | -------- | -----------
| `vm_host` | `string` | `true` | The `id` or `name` of an existing MAAS VM host.
| `cores` | `int` | `false` | The number of CPU cores (defaults to `1`).
| `pinned_cores` | `int` | `false` | List of host CPU cores to pin the VM to. If this is passed, the `cores` parameter is ignored.
| `memory` | `int` | `false` | The amount of memory, specified in MiB.
| `storage` | `string` | `false` | A list of storage constraint identifiers in the form `label:size(tag,tag,...),label:size(tag,tag,...)`. For more information, see [this](https://maas.io/docs/composable-hardware#heading--storage).
| `interfaces` | `string` | `false` | A labeled constraint map associating constraint labels with desired interface properties. MAAS will assign interfaces that match the given interface properties. For more information, see [this](https://maas.io/docs/composable-hardware#heading--interfaces).
| `hostname` | `string` | `false` | The hostname of the newly composed machine.
| `domain` | `string` | `false` | The name of the domain in which to put the newly composed machine.
| `zone` | `string` | `false` | The name of the zone in which to put the newly composed machine.
| `pool` | `string` | `false` | The name of the pool in which to put the newly composed machine.
