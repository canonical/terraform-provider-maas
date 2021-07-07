
# Resource: maas_fabric

Provides a resource to manage MAAS network fabrics.

## Example Usage

```terraform
resource "maas_fabric" "tf_fabric" {
  name = "tf-fabric"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The fabric name.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The fabric ID.

## Import

An existing network fabric can be imported using its name or ID. e.g.

```shell
terraform import maas_fabric.tf_fabric tf-fabric
```
