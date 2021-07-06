
# Resource: maas_block_device

Provides a resource to manage MAAS machines' block devices.

## Example Usage

```terraform
resource "maas_block_device" "vdb" {
  machine = maas_machine.virsh_vm2.id
  name = "vdb"
  id_path = "/dev/vdb"
  size_gigabytes = 27
  tags = [
    "ssd",
  ]

  partitions {
    size_gigabytes = 10
    fs_type = "ext4"
    label = "media"
    mount_point = "/media"
  }

  partitions {
    size_gigabytes = 15
    fs_type = "ext4"
    mount_point = "/storage"
  }
}
```

## Argument Reference

The following arguments are supported:

* `machine` - (Required) The machine identifier (system ID, hostname, or FQDN) that owns the block device.
* `name` - (Required) The block device name.
* `size_gigabytes` - (Required) The size of the block device (given in GB).
* `block_size` - (Optional) The block size of the block device. Defaults to `512`.
* `is_boot_device` - (Optional) Boolean value indicating if the block device is set as the boot device.
* `partitions` - (Optional) List of partition resources created for the new block device. Parameters defined below. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html). And, it is computed if it's not given.
* `model` - (Optional) Model of the block device. Used in conjunction with `serial` argument. Conflicts with `id_path`. This argument is computed if it's not given.
* `serial` - (Optional) Serial number of the block device. Used in conjunction with `model` argument. Conflicts with `id_path`. This argument is computed if it's not given.
* `id_path` - (Optional) Only used if `model` and `serial` cannot be provided. This should be a path that is fixed and doesn't change depending on the boot order or kernel version. This argument is computed if it's not given.
* `tags` - (Optional) A set of tag names assigned to the new block device. This argument is computed if it's not given.

### partitions

* `size_gigabytes` - (Required) The partition size (given in GB).
* `bootable` - (Optional) Boolean value indicating if the partition is set as bootable.
* `tags` - (Optional) The tags assigned to the new block device partition.
* `fs_type` - (Optional) The file system type (e.g. `ext4`). If this is not set, the partition is unformatted.
* `label` - (Optional) The label assigned if the partition is formatted.
* `mount_point` - (Optional) The mount point used. If this is not set, the partition is not mounted. This is used only the partition is formatted.
* `mount_options` - (Optional) The options used for the partition mount.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - Block device ID.
* `uuid` - Block device UUID.
* `path` - Block device path.

## Import

Block devices can be imported with the machine identifier (system ID, hostname, or FQDN) and the block device identifier (ID or name). e.g.

```shell
terraform import maas_block_device.vdb machine-06:vdb
```
