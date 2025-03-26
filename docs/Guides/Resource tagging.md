---
# generated using the Resource tagging.md.tmpl template
page_title: "Resource tagging - terraform-provider-maas"
subcategory: ""
description: |-
  MAAS implements tags in a number of ways to help manage and configure entities. This guide provides an overview of how to manage tags in MAAS using the terraform-provider-maas provider. Note that not all resources support tags.
---

# Resource tagging

MAAS implements tags in a number of ways to help manage and configure entities. This guide provides an overview of how to manage tags in MAAS using the terraform-provider-maas provider. Note that not all resources support tags.

## Machine tags

The `maas_tag` resource is used to manage Machine tags in MAAS. These tags can contain information such as kernel options and an XPATH definition that defines automatic tagging of machines. These tags cannot be used with other resources such as network interfaces or block devices. 

`maas_instance.tags` refers to the name of these tags.

See the [maas_tag](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/tag.md) resource for more information.

## Other tags

Other MAAS entities use simple strings as tags. They can be managed using the `tags` attribute on the relevant resource:

```nohighlight
resource "maas_network_interface_physical" "example" {
  ...
  tags = ["one", "two"]
}
```

Some examples of resources that support tags are provided below:

- [maas_network_interface_physical](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/network_interface_physical.md)
- [maas_network_interface_bridge](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/network_interface_bridge.md)
- [maas_network_interface_vlan](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/network_interface_vlan.md)
- [maas_block_device](https://github.com/maas/terraform-provider-maas/blob/master/docs/resources/block_device.md)

### Managing tags with dedicated resources

Where possible, it's recommended to use the `tags` attribute on the relevant MAAS resources, such as `maas_network_interface_physical.tags`, to manage tags. Additional, dedicated tag resources such as `maas_network_interface_tag` are provided for cases where tags need to be managed separately from the specific resource, for example when chaining 2 terraform modules together to separate configuration and deployment, or when managing tags on resources that are not managed by Terraform.

> [!WARNING] 
> Using dedicated tag resources together with the `tags` attribute on MAAS specific resources, such as `maas_network_interface_physical.tags`, will cause conflicts and will overwrite the tags already set.

If it's desirable to create interfaces using the `tags` attribute and later manage tags using the relevant dedicated resource, the `[ignore_changes](https://developer.hashicorp.com/terraform/language/meta-arguments/lifecycle#ignore_changes)` lifecycle argument can be used to ignore changes to a particular resource after creation

```nohighlight
resource "maas_network_interface_physical" "example" {
  # ...
  tags = ["one", "two"]
  lifecycle {
    ignore_changes = [
    # ignore changes to tags after creation because they will be managed with the dedicated `maas_network_interface_tag` resource later on.
    tags
    ]
  }
}
```
