
# Resource: maas_subnet_ip_range

Provides a resource to manage MAAS network subnets IP ranges.

## Example Usage

```terraform
resource "maas_subnet_ip_range" "dynamic_ip_range" {
  subnet = maas_subnet.tf_subnet_2.id
  type = "dynamic"
  start_ip = "10.77.77.2"
  end_ip = "10.77.77.60"
}
```

## Argument Reference

The following arguments are supported:

* `subnet` - (Required) The subnet identifier (ID or CIDR) for the new IP range.
* `type` - (Required) The IP range type. Valid options are: `dynamic`, `reserved`.
* `start_ip` - (Required) The start IP for the new IP range (inclusive).
* `end_ip` - (Required) The end IP for the new IP range (inclusive).
* `comment` - (Optional) A description of this range. This argument is computed if it's not set.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The subnet IP range ID.

## Import

IP ranges can be imported with the start IP and the end IP. e.g.

```shell
terraform import maas_subnet_ip_range.dynamic_ip_range 10.77.77.2:10.77.77.60
```
