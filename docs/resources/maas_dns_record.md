
# Resource: maas_dns_record

Provides a resource to manage MAAS DNS domain records.

## Example Usage

```terraform
resource "maas_dns_record" "test_a" {
  type = "A/AAAA"
  data = "10.99.11.33"
  fqdn = "test-a.${maas_dns_domain.cloudbase.name}"
}
```

## Argument Reference

The following arguments are supported:

* `type` - (Required) The DNS record type. Valid options are: `A/AAAA`, `CNAME`, `MX`, `NS`, `SRV`, `SSHFP`, `TXT`.
* `data` - (Required) The data set for the new DNS record.
* `name` - (Optional) The new DNS record resource name. Used in conjunction with `domain`. It conflicts with `fqdn` argument.
* `domain` - (Optional) The domain of the new DNS record. Used in conjunction with `name`. It conflicts with `fqdn` argument.
* `fqdn` - (Optional) The fully qualified domain name of the new DNS record. This contains the name and the domain of the new DNS record. It conflicts with `name` and `domain` arguments.
* `ttl` - (Optional) The TTL of the new DNS record.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The DNS record ID.

## Import

DNS records can be imported using the type and the identifier (ID or FQDN). e.g.

```shell
terraform import maas_dns_record.test_a A/AAAA:test-a.cloudbase
```
