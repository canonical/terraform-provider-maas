
# Resource: maas_instance

Provides a resource to deploy and release machines already configured in MAAS, based on the specified parameters. If no parameters are given, a random machine will be allocated and deployed using the defaults.

**NOTE:** The MAAS provider currently provides both standalone resources and in-line resources for network interfaces. You cannot use in-line network interfaces in conjunction with any standalone network interfaces resources. Doing so will cause conflicts and will overwrite network configs.

## Example Usage

```terraform
resource "maas_instance" "two_random_nodes_2G" {
  count = 2
  allocate_params {
    min_cpu_count = 1
    min_memory = 2048
  }
  deploy_params {
    distro_series = "focal"
  }
}
```

## Argument Reference

The following arguments are supported:

* `allocate_params` - (Optional) Nested argument with the constraints used to machine allocation. Defined below.
* `deploy_params` - (Optional) Nested argument with the config used to deploy the allocated machine. Defined below.
* `network_interfaces` - (Optional) Specifies a network interface configuration done before the machine is deployed. Parameters defined below. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).

### allocate_params

* `min_cpu_count` - (Optional) The minimum number of cores used to allocate the MAAS machine.
* `min_memory` - (Optional) The minimum RAM memory size (in MB) used to allocate the MAAS machine.
* `hostname` - (Optional) The hostname of the MAAS machine to be allocated.
* `zone` - (Optional) The zone name of the MAAS machine to be allocated.
* `pool` - (Optional) The pool name of the MAAS machine to be allocated.
* `tags` - (Optional) A set of tag names that must be assigned on the MAAS machine to be allocated.

### deploy_params

* `distro_series` - (Optional) The distro series used to deploy the allocated MAAS machine. If it's not given, the MAAS server default value is used.
* `hwe_kernel` - (Optional) Hardware enablement kernel to use with the image. Only used when deploying Ubuntu.
* `user_data` - (Optional) Cloud-init user data script that gets run on the machine once it has deployed. A good practice is to set this with `file("/tmp/user-data.txt")`, where `/tmp/user-data.txt` is a cloud-init script.

### network_interfaces

* `name` - (Required) The name of the network interface to be configured on the allocated machine.
* `subnet_cidr` - (Optional) An existing subnet CIDR used to configure the network interface. Unless `ip_address` is defined, a free IP address is allocated from the subnet.
* `ip_address` - (Optional) Static IP address to be configured on the network interface. If this is set, the `subnet_cidr` is required.

**NOTE:** If both `subnet_cidr` and `ip_address` are not defined, the interface will not be configured on the allocated machine.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The deployed MAAS machine system ID.
* `fqdn` - The deployed MAAS machine FQDN.
* `hostname` - The deployed MAAS machine hostname.
* `zone` - The deployed MAAS machine zone name.
* `pool` - The deployed MAAS machine pool name.
* `tags` - A set of tag names associated to the deployed MAAS machine.
* `cpu_count` - The number of CPU cores of the deployed MAAS machine.
* `memory` -  The RAM memory size (in GiB) of the deployed MAAS machine.
* `ip_addresses` - A set of IP addressed assigned to the deployed MAAS machine.

## Import

The machines imported as `maas_instance` resources must be already deployed. They can be imported using one of the deployed machine attributes: system ID, hostname, or FQDN. e.g.

```shell
terraform import maas_instance.virsh_vm machine-01
```
