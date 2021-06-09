# `maas_vlan`

Get an existing MAAS VLAN.

Example:

```hcl
data "maas_fabric" "default" {
  name = "maas"
}

data "maas_vlan" "default" {
  fabric_id = data.maas_fabric.default.id
  vid = 0
}
```

Parameters:

| Name | Type | Required | Description
| ---- | ---- | -------- | -----------
| `fabric_id` | `int` | `true` | The ID of the fabric containing the VLAN.
| `vid` | `int` | `true` | The VLAN traffic segregation ID.
