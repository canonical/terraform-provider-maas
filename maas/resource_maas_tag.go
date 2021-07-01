package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

func resourceMaasTag() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTagCreate,
		ReadContext:   resourceTagRead,
		UpdateContext: resourceTagUpdate,
		DeleteContext: resourceTagDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				client := m.(*client.Client)
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"machines": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceTagCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

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

	return resourceTagUpdate(ctx, d, m)
}

func resourceTagRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	if _, err := client.Tag.Get(d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceTagUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

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

	return resourceTagRead(ctx, d, m)
}

func resourceTagDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	if err := client.Tag.Delete(d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getTagCreateParams(d *schema.ResourceData) *entity.TagParams {
	return &entity.TagParams{
		Name: d.Get("name").(string),
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
	machinesSystemIDs := []string{}
	for _, machineIdentifier := range convertToStringSlice(p.(*schema.Set).List()) {
		machine, err := getMachine(client, machineIdentifier)
		if err != nil {
			return nil, err
		}
		machinesSystemIDs = append(machinesSystemIDs, machine.SystemID)
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
