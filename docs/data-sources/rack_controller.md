---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "maas_rack_controller Data Source - terraform-provider-maas"
subcategory: ""
description: |-
  Provides details about an existing MAAS rack controller.
---

# maas_rack_controller (Data Source)

Provides details about an existing MAAS rack controller.

## Example Usage

```terraform
data "maas_rack_controller" "test_rack_controller" {
  hostname = "maas-rack-0"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `hostname` (String) The hostname of the rack controller.

### Read-Only

- `description` (String) The description of the rack controller.
- `id` (String) The ID of this resource.
- `services` (Set of Object) The services running on the rack controller. (see [below for nested schema](#nestedatt--services))
- `version` (String) The MAAS version of the rack controller.

<a id="nestedatt--services"></a>
### Nested Schema for `services`

Read-Only:

- `name` (String)
- `status` (String)