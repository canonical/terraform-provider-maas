package maas

import (
	"context"
	"fmt"
	"slices"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMaasTag() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage a MAAS tag.",
		CreateContext: resourceTagCreate,
		ReadContext:   resourceTagRead,
		UpdateContext: resourceTagUpdate,
		DeleteContext: resourceTagDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client

				tag, err := getTag(client, d.Id())
				if err != nil {
					return nil, err
				}
				machines, err := client.Tag.GetMachines(tag.Name)
				if err != nil {
					return nil, err
				}
				machinesSystemIDs := make([]string, len(machines))
				for i, machine := range machines {
					machinesSystemIDs[i] = machine.SystemID
				}
				tfState := map[string]interface{}{
					"id":       tag.Name,
					"name":     tag.Name,
					"machines": machinesSystemIDs,
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"comment": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "A description of what the the tag will be used for in natural language.",
			},
			"definition": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An XPATH query that is evaluated against the hardware_details stored for all nodes. (i.e. the output of ``lshw -xml``)",
			},
			"kernel_opts": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Nodes associated with this tag will add this string to their kernel options when booting. The value overrides the global ``kernel_opts`` setting. If more than one tag is associated with a node, command line will be concatenated from all associated tags, in alphabetic tag name order.",
			},
			"machines": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "List of MAAS machines' identifiers (system ID, hostname, or FQDN) that will be tagged with the new tag.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The new tag name. Because the name will be used in urls, it should be short.",
			},
		},
	}
}

func resourceTagCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	params := getTagCreateParams(d)
	tag, err := findTag(client, params.Name)
	if err != nil {
		return diag.FromErr(err)
	}
	if tag == nil {
		tag, err = client.Tags.Create(params)
		if err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(tag.Name)

	return resourceTagUpdate(ctx, d, meta)
}

func resourceTagRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	tag, err := findTag(client, d.Id())
	if err != nil {
		return diag.FromErr(err)
	} else if tag == nil {
		d.SetId("")
		return nil
	}

	d.Set("definition", tag.Definition)
	d.Set("comment", tag.Comment)
	d.Set("kernel_opts", tag.KernelOpts)

	return nil
}

func resourceTagUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	if d.HasChanges("definition", "comment", "kernel_opts") {
		if _, err := client.Tag.Update(d.Id(), getTagCreateParams(d)); err != nil {
			return diag.FromErr(err)
		}
	}

	tagMachinesIDs, err := getTagTFMachinesSystemIDs(client, d)
	if err != nil {
		return diag.FromErr(err)
	}
	if len(tagMachinesIDs) > 0 {
		// Tag specified machines
		err := client.Tag.AddMachines(d.Id(), tagMachinesIDs)
		if err != nil {
			return diag.FromErr(err)
		}
		// Untag previously tagged machines
		err = untagOtherMachines(client, d.Id(), tagMachinesIDs)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceTagRead(ctx, d, meta)
}

func resourceTagDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	if err := client.Tag.Delete(d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getTagCreateParams(d *schema.ResourceData) *entity.TagParams {
	return &entity.TagParams{
		Name:       d.Get("name").(string),
		Definition: d.Get("definition").(string),
		Comment:    d.Get("comment").(string),
		KernelOpts: d.Get("kernel_opts").(string),
	}
}

func findTag(client *client.Client, tagName string) (*entity.Tag, error) {
	tags, err := client.Tags.Get()
	if err != nil {
		return nil, err
	}
	for _, t := range tags {
		if t.Name == tagName {
			return &t, nil
		}
	}
	return nil, nil
}

func getTag(client *client.Client, tagName string) (*entity.Tag, error) {
	tag, err := findTag(client, tagName)
	if err != nil {
		return nil, err
	}
	if tag == nil {
		return nil, fmt.Errorf("tag (%s) was not found", tagName)
	}
	return tag, nil
}

func getTagTFMachinesSystemIDs(client *client.Client, d *schema.ResourceData) ([]string, error) {
	p, ok := d.GetOk("machines")
	if !ok {
		return nil, nil
	}
	machines, err := client.Machines.Get(&entity.MachinesParams{})
	if err != nil {
		return nil, err
	}
	machinesSystemIDs := []string{}
	for _, identifier := range convertToStringSlice(p.(*schema.Set).List()) {
		found := false
		for _, m := range machines {
			if slices.Contains([]string{m.SystemID, m.Hostname, m.FQDN}, identifier) {
				if slices.Contains(machinesSystemIDs, m.SystemID) {
					return nil, fmt.Errorf("machine (%s) is referenced more than once", m.SystemID)
				}
				machinesSystemIDs = append(machinesSystemIDs, m.SystemID)
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("machine (%s) not found", identifier)
		}
	}

	return machinesSystemIDs, nil
}

func untagOtherMachines(client *client.Client, tagName string, taggedMachineIDs []string) error {
	machines, err := client.Tag.GetMachines(tagName)
	if err != nil {
		return err
	}
	otherMachines := []string{}
	for _, m := range machines {
		found := false
		for _, id := range taggedMachineIDs {
			if m.SystemID == id {
				found = true
				break
			}
		}
		if found {
			continue
		}
		otherMachines = append(otherMachines, m.SystemID)
	}
	if len(otherMachines) > 0 {
		client.Tag.RemoveMachines(tagName, otherMachines)
	}
	return nil
}
