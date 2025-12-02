package maas

import (
	"context"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASNetworkInterfaceTag() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage tags as strings on a network interface.",
		CreateContext: resourceNetworkInterfaceTagCreate,
		ReadContext:   resourceNetworkInterfaceTagRead,
		UpdateContext: resourceNetworkInterfaceTagUpdate,
		DeleteContext: resourceNetworkInterfaceTagDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client
				// Get the system ID and interface ID from the user inputted resource ID
				systemID, interfaceID, err := SplitTagStateID(d.Id())
				if err != nil {
					return nil, err
				}
				// Gets the existing interface, thereby ensuring that the node and interface exists.
				existingInterface, err := client.NetworkInterface.Get(systemID, interfaceID)
				if err != nil {
					return nil, err
				}
				// Infer the type of the relevant entity from the system ID. Makes calls to MAAS.
				entityType, err := getMachineOrDeviceTypeFromSystemID(client, systemID)
				if err != nil {
					return nil, err
				}
				// Set the machine or device in state, set once.
				if entityType == "machine" {
					if err := d.Set("machine", existingInterface.SystemID); err != nil {
						return nil, err
					}
				} else {
					if err := d.Set("device", existingInterface.SystemID); err != nil {
						return nil, err
					}
				}
				// Set the resource ID
				d.SetId(fmt.Sprintf("%v/%v", existingInterface.SystemID, existingInterface.ID))

				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"device": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"machine", "device"},
				Description:  "The identifier (system ID, hostname, or FQDN) of the device with the network interface. Either `machine` or `device` must be provided.",
			},
			"interface_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The network interface ID to tag.",
			},
			"machine": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"machine", "device"},
				Description:  "The identifier (system ID, hostname, or FQDN) of the machine with the network interface. Either `machine` or `device` must be provided.",
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

func resourceNetworkInterfaceTagCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	interfaceID := d.Get("interface_id").(int)
	desiredTags := convertToStringSlice(d.Get("tags").(*schema.Set).List())

	systemID, err := getMachineOrDeviceSystemID(client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the existing interface
	existingInterface, err := client.NetworkInterface.Get(systemID, interfaceID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Remove tags that are not in the desired set
	if existingInterface.Tags == nil {
		existingInterface.Tags = []string{}
	}

	for _, tag := range existingInterface.Tags {
		if !slices.Contains(desiredTags, tag) {
			_, err := client.NetworkInterface.RemoveTag(systemID, interfaceID, tag)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// Add tags that are in the desired set. AddTag will not add duplicates.
	for _, tag := range desiredTags {
		if !slices.Contains(existingInterface.Tags, tag) {
			_, err := client.NetworkInterface.AddTag(systemID, interfaceID, tag)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// Create the resource ID in state. A unique resource for every interface.
	d.SetId(fmt.Sprintf("%v/%v", systemID, interfaceID))

	// Read the resource to update state
	return resourceNetworkInterfaceTagRead(ctx, d, meta)
}

func resourceNetworkInterfaceTagRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	systemID, interfaceID, err := SplitTagStateID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	// Get the existing interface
	existingInterface, err := client.NetworkInterface.Get(systemID, interfaceID)
	if err != nil {
		return diag.FromErr(err)
	}
	// Set the tags in state
	if err := d.Set("tags", existingInterface.Tags); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("interface_id", interfaceID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkInterfaceTagUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	systemID, err := getMachineOrDeviceSystemID(client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	interfaceID := d.Get("interface_id").(int)
	// Get the existing interface
	existingInterface, err := client.NetworkInterface.Get(systemID, interfaceID)
	if err != nil {
		return diag.FromErr(err)
	}

	existingTags := existingInterface.Tags
	desiredTags := convertToStringSlice(d.Get("tags").(*schema.Set).List())

	// Remove tags that are not in the specified set
	for _, tag := range existingTags {
		if !slices.Contains(desiredTags, tag) {
			_, err := client.NetworkInterface.RemoveTag(systemID, interfaceID, tag)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	// Add tags that are in the specified set
	for _, tag := range desiredTags {
		if !slices.Contains(existingTags, tag) {
			_, err := client.NetworkInterface.AddTag(systemID, interfaceID, tag)
			if err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceNetworkInterfaceTagRead(ctx, d, meta)
}

func resourceNetworkInterfaceTagDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	systemID, err := getMachineOrDeviceSystemID(client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	interfaceID := d.Get("interface_id").(int)
	tags := convertToStringSlice(d.Get("tags").(*schema.Set).List())

	for _, t := range tags {
		_, err := client.NetworkInterface.RemoveTag(systemID, interfaceID, t)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId("")

	return nil
}

// SplitTagStateID splits the state ID of a tag in the format system_id:interface_id into its component ids, where system_id is the system ID of the machine or device, and interface_id is the ID of the network interface.
func SplitTagStateID(stateID string) (string, int, error) {
	splitID := strings.SplitN(stateID, "/", 2)
	if len(splitID) != 2 {
		return "", 0, fmt.Errorf("invalid resource ID: %s", stateID)
	}

	interfaceID, err := strconv.Atoi(splitID[1])
	if err != nil {
		return "", 0, err
	}

	return splitID[0], interfaceID, nil
}
