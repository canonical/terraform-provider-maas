
# Resource: maas_space

Provides a resource to manage MAAS network spaces.

## Example Usage

```terraform
resource "maas_space" "tf_space" {
  name = "tf-space"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the new space.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The space ID.

## Import

Spaces can be imported using the name or ID. e.g.

```shell
terraform import maas_space.tf_space tf-space
```
