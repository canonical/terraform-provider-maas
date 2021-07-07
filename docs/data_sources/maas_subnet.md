# Data Source: maas_subnet

Provides details about an existing MAAS network subnet.

## Example Usage

```terraform
data "maas_subnet" "vid10" {
  cidr = "10.10.0.0/16"
}
```

## Argument Reference

The following arguments are supported:

* `cidr` - (Requried) The subnet CIDR.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The subnet ID.
* `fabric` - The subnet fabric.
* `vid` - The subnet VLAN traffic segregation ID.
* `name` - The subnet name.
* `rdns_mode` - How reverse DNS is handled for this subnet. It can have one of the following values:
  * `0` - Disabled, no reverse zone is created.
  * `1` - Enabled, generate reverse zone.
  * `2` - RFC2317, extends `1` to create the necessary parent zone with the appropriate CNAME resource records for the network, if the network is small enough to require the support described in RFC2317.
* `allow_dns` - Boolean value that indicates if the MAAS DNS resolution is enabled for this subnet.
* `allow_proxy` - Boolean value that indicates if `maas-proxy` allows requests from this subnet.
* `gateway_ip` - Gateway IP address for the subnet.
* `dns_servers` - List of IP addresses set as DNS servers for the subnet.
