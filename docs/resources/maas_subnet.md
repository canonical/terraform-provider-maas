
# Resource: maas_subnet

Provides a resource to manage MAAS network subnets.

**NOTE:** The MAAS provider currently supports both standalone resources and in-line resources for subnet IP ranges. You cannot use in-line `ip_ranges` in conjunction with standalone `maas_subnet_ip_range` resources. Doing so will cause conflicts and will overwrite subnet IP ranges.

## Example Usage

```terraform
resource "maas_subnet" "tf_subnet" {
  cidr = "10.88.88.0/24"
  fabric = maas_fabric.tf_fabric.id
  vlan = maas_vlan.tf_vlan.vid
  name = "tf_subnet"
  gateway_ip = "10.88.88.1"
  dns_servers = [
    "1.1.1.1",
  ]

  ip_ranges {
    type = "reserved"
    start_ip = "10.88.88.1"
    end_ip = "10.88.88.50"
  }
  ip_ranges {
    type = "dynamic"
    start_ip = "10.88.88.200"
    end_ip = "10.88.88.254"
  }
}
```

## Argument Reference

The following arguments are supported:

* `cidr` - (Required) The subnet CIDR.
* `name` - (Optional) The subnet name.
* `fabric` - (Optional) The fabric identifier (ID or name) for the new subnet.
* `vlan` - (Optional) The VLAN identifier (ID or traffic segregation ID) for the new subnet. If this is set, the `fabric` argument is required.
* `ip_ranges` - (Optional) A set of IP ranges configured on the new subnet. Parameters defined below. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).
* `rdns_mode` - (Optional) How reverse DNS is handled for this subnet. Defaults to `2`. Valid options are:
  * `0` - Disabled, no reverse zone is created.
  * `1` - Enabled, generate reverse zone.
  * `2` - RFC2317, extends `1` to create the necessary parent zone with the appropriate CNAME resource records for the network, if the network is small enough to require the support described in RFC2317.
* `allow_dns` - (Optional) Boolean value that indicates if the MAAS DNS resolution is enabled for this subnet. Defaults to `true`.
* `allow_proxy` - (Optional) Boolean value that indicates if `maas-proxy` allows requests from this subnet. Defaults to `true`.
* `gateway_ip` - (Optional) Gateway IP address for the new subnet. This argument is computed if it's not set.
* `dns_servers` - (Optional) List of IP addresses set as DNS servers for the new subnet. This argument is computed if it's not set.

### ip_ranges

* `type` - (Required) The IP range type. Valid options are: `dynamic`, `reserved`.
* `start_ip` - (Required) The start IP for the new IP range (inclusive).
* `end_ip` - (Required) The end IP for the new IP range (inclusive).
* `comment` - (Optional) A description of this range.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The subnet ID.

## Import

MAAS network subnets can be imported using the ID or CIDR. e.g.

```shell
terraform import maas_subnet.tf_subnet 10.77.77.0/24
```
