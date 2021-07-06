# Data Source: maas_fabric

Provides details about an existing MAAS network fabric.

## Example Usage

```terraform
data "maas_fabric" "default" {
  name = "maas"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The fabric name.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The fabric ID.
