# `maas_fabric`

Get an existing MAAS fabric.

Example:

```hcl
data "maas_fabric" "default" {
  name = "maas"
}
```

Parameters:

| Name | Type | Required | Description
| ---- | ---- | -------- | -----------
| `name` | `string` | `true` | The fabric name.
