package maas

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMaasBlockDeviceTag() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage tags as strings on a block device.",
		CreateContext: resourceBlockDeviceTagCreate,
		ReadContext:   resourceBlockDeviceTagRead,
		UpdateContext: resourceBlockDeviceTagUpdate,
		DeleteContext: resourceBlockDeviceTagDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				systemID, blockDeviceID, err := SplitTagStateId(d.Id())
				if err != nil {
					return nil, err
				}
				client := meta.(*ClientConfig).Client
				blockDevice, err := client.BlockDevice.Get(systemID, blockDeviceID)
				if err != nil {
					return nil, err
				}

				d.SetId(fmt.Sprintf("%v/%v", blockDevice.SystemID, blockDevice.ID))
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"block_device_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The block device ID to tag.",
			},
			"machine": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The identifier (system ID, hostname, or FQDN) of the machine with the network interface.",
			},
			"tags": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "The tags to assign to the network interface. Tag names should be short and without spaces.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceBlockDeviceTagCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client
	machineId := d.Get("machine").(string)
	machine, err := getMachine(client, machineId)
	if err != nil {
		return diag.FromErr(err)
	}
	blockDeviceId := d.Get("block_device_id").(int)

	// Get the block device
	blockDevice, err := getBlockDevice(client, machine.SystemID, fmt.Sprintf("%v", blockDeviceId))
	if err != nil {
		return diag.FromErr(err)
	}

	// Remove existing tags
	desiredTags := convertToStringSlice(d.Get("tags").(*schema.Set).List())
	existingTags := blockDevice.Tags
	for _, tag := range existingTags {
		if !slices.Contains(desiredTags, tag) {
			_, err := client.BlockDevice.RemoveTag(machine.SystemID, blockDevice.ID, tag)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// Add new tags
	for _, tag := range desiredTags {
		if !slices.Contains(existingTags, tag) {
			_, err := client.BlockDevice.AddTag(machine.SystemID, blockDevice.ID, tag)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	d.SetId(fmt.Sprintf("%v/%v", machine.SystemID, blockDevice.ID))

	return nil
}

func resourceBlockDeviceTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Get the existing block device
	systemID, blockDeviceID, err := SplitTagStateId(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	blockDevice, err := client.BlockDevice.Get(systemID, blockDeviceID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the attributes in state
	tfstate := map[string]interface{}{
		"tags":            blockDevice.Tags,
		"machine":         blockDevice.SystemID,
		"block_device_id": blockDevice.ID,
	}
	if err := setTerraformState(d, tfstate); err != nil {
		return diag.FromErr(fmt.Errorf("could not set block device properties: %v", err))
	}
	return nil
}

func resourceBlockDeviceTagUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	systemID := d.Get("machine").(string)
	blockDeviceID := d.Get("block_device_id").(int)
	blockDevice, err := client.BlockDevice.Get(systemID, blockDeviceID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Remove undesired tags
	desiredTags := convertToStringSlice(d.Get("tags").(*schema.Set).List())
	existingTags := blockDevice.Tags
	for _, tag := range existingTags {
		if !slices.Contains(desiredTags, tag) {
			_, err := client.BlockDevice.RemoveTag(blockDevice.SystemID, blockDevice.ID, tag)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// Add new tags
	for _, tag := range desiredTags {
		if !slices.Contains(existingTags, tag) {
			_, err := client.BlockDevice.AddTag(blockDevice.SystemID, blockDevice.ID, tag)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}
	return resourceBlockDeviceTagRead(ctx, d, meta)
}

func resourceBlockDeviceTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	systemID := d.Get("machine").(string)
	blockDeviceID := d.Get("block_device_id").(int)

	// Remove all tags specified
	desiredTags := convertToStringSlice(d.Get("tags").(*schema.Set).List())
	for _, tag := range desiredTags {
		_, err := client.BlockDevice.RemoveTag(systemID, blockDeviceID, tag)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId("")
	return nil
}
