
# `maas_tag`

Create a new MAAS tag, and use it to tag MAAS machines.

Example:

```hcl
resource "maas_tag" "kvm" {
  name = "kvm"
  machine_ids = [
    maas_vm_host_machine.kvm[0].id,
    maas_vm_host_machine.kvm[1].id,
    maas_machine.virsh_vm1.id,
    maas_machine.virsh_vm2.id,
  ]
}
```

Parameters:

| Name | Type | Required | Description
| ---- | ---- | -------- | -----------
| `name` | `string` | `true` | The new tag name. Because the name will be used in urls, it should be short.
| `machine_ids` | `[]string` | `false` | List of MAAS machines' ids that will be tagged.
