package maas

import (
	"context"

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
				if err := d.Set("name", d.Id()); err != nil {
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
			"machine_ids": {
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

	_, err := client.Tag.Get(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceTagUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	if p, ok := d.GetOk("machine_ids"); ok {
		machineIds := convertToStringSlice(p.(*schema.Set).List())
		// Tag specified machines
		err := client.Tag.AddMachines(d.Id(), machineIds)
		if err != nil {
			return diag.FromErr(err)
		}
		// Untag previously tagged machines
		err = untagOtherMachines(client, d.Id(), machineIds)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceTagRead(ctx, d, m)
}

func resourceTagDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	err := client.Tag.Delete(d.Id())
	if err != nil {
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

func untagOtherMachines(client *client.Client, tagName string, taggedMachineIds []string) error {
	machines, err := client.Tag.GetMachines(tagName)
	if err != nil {
		return err
	}

	otherMachines := []string{}
	for _, m := range machines {
		found := false
		for _, id := range taggedMachineIds {
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
