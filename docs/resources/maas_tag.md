
# Resource: maas_tag

Provides a resource to manage a MAAS tag.

## Example Usage

```terraform
resource "maas_tag" "kvm" {
  name = "kvm"
  machines = [
    maas_vm_host_machine.kvm[0].id,
    maas_vm_host_machine.kvm[1].id,
    maas_machine.virsh_vm1.id,
    maas_machine.virsh_vm2.id,
  ]
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The new tag name. Because the name will be used in urls, it should be short.
* `machines` - (Optional) List of MAAS machines' identifiers (system ID, hostname, or FQDN) that will be tagged with the new tag.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The tag name.

## Import

An existing tag can be imported using its name. e.g.

```shell
terraform import maas_tag.kvm kvm
```
