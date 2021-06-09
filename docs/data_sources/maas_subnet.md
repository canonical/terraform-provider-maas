# `maas_subnet`

Fetches details about an existing MAAS subnet.

Example:

```hcl
data "maas_fabric" "default" {
  name = "maas"
}

data "maas_vlan" "default" {
  fabric_id = data.maas_fabric.default.id
  vid = 0
}

data "maas_subnet" "pxe" {
  cidr = "10.121.0.0/16"
  vlan_id = data.maas_vlan.default.id
}
```

Parameters:

| Name | Type | Required | Description
| ---- | ---- | -------- | -----------
| `cidr` | `string` | `true` | The network CIDR for this subnet.
| `vlan_id` | `int` | `false` | The ID of the VLAN this subnet belongs to. This is the unique identifier set by MAAS for the VLAN resource, not the actual VLAN traffic segregation ID.
