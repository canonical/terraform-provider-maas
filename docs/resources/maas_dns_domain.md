
# Resource: maas_dns_domain

Provides a resource to manage MAAS DNS domains.

## Example Usage

```terraform
resource "maas_dns_domain" "cloudbase" {
  name = "cloudbase"
  ttl = 3600
  authoritative = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the new DNS domain.
* `ttl` - (Optional) The default TTL for the new DNS domain.
* `authoritative` - (Optional) Boolean value indicating if the new DNS domain is authoritative. Defaults to `false`.
* `is_default` - (Optional) Boolean value indicating if the new DNS domain will be set as the default in the MAAS environment. Defaults to `false`.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The DNS domain ID.

## Import

DNS domains can be imported using their ID or name. e.g.

```shell
terraform import maas_dns_domain.cloudbase cloudbase
```
