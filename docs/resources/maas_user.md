
# Resource: maas_user

Provides a resource to manage MAAS users.

## Example Usage

```terraform
resource "maas_user" "cloudbase" {
  name = "cloudbase"
  password = "Passw0rd123"
  email = "admin@cloudbase.local"
  is_admin = true
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The user name.
* `password` - (Required) The user password.
* `email` - (Required) The user e-mail address.
* `is_admin` - (Optional) Boolean value indicating if the user is a MAAS administrator. Defaults to `false`.

## Import

Users can be imported using their name. e.g.

```shell
terraform import maas_user.cloudbase cloudbase
```
