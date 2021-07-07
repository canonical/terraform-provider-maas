
# Resource: maas_vm_host_machine

Provides a resource to manage MAAS VM host machines.

## Example Usage

```terraform
resource "maas_vm_host_machine" "kvm" {
  vm_host = maas_vm_host.kvm.id
  cores = 1
  memory = 2048

  network_interfaces {
    name = "eth0"
    subnet_cidr = data.maas_subnet.pxe.cidr
  }

  storage_disks {
    size_gigabytes = 10
  }
  storage_disks {
    size_gigabytes = 15
  }
}
```

## Argument Reference

The following arguments are supported:

* `vm_host` - (Required) ID or name of the VM host used to compose the new machine.
* `cores` - (Optional) The number of CPU cores (defaults to 1).
* `pinned_cores` - (Optional) List of host CPU cores to pin the VM host machine to. If this is passed, the `cores` parameter is ignored.
* `memory` - (Optional) The VM host machine RAM memory, specified in MB (defaults to 2048).
* `network_interfaces` - (Optional) A list of network interfaces for new the VM host. This argument only works when the VM host is deployed from a registered MAAS machine. Parameters defined below. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).
* `storage_disks` - (Optional) A list of storage disks for the new VM host. Parameters defined below. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).
* `hostname` - (Optional) The VM host machine hostname. This is computed if it's not set.
* `domain` - (Optional) The VM host machine domain. This is computed if it's not set.
* `zone` - (Optional) The VM host machine zone. This is computed if it's not set.
* `pool` - (Optional) The VM host machine pool. This is computed if it's not set.

### network_interfaces

* `name` - (Required) The network interface name.
* `fabric` - (Optional) The fabric for the network interface.
* `vlan` - (Optional) The VLAN for the network interface.
* `subnet_cidr` - (Optional) The subnet CIDR for the network interface.
* `ip_address` - (Optional) Static IP configured on the new network interface.

### storage_disks

* `size_gigabytes` - (Required) The storage disk size, specified in GB.
* `pool` - (Optional) The VM host storage pool name.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The VM host machine system ID.

## Import

VM host machines can be imported using the identifier of the MAAS machine (system ID, hostname, or FQDN). e.g.

```shell
terraform import maas_vm_host_machine.test machine-02
```
