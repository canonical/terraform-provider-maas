package maas

import (
	"context"
	"fmt"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMaasBootSourceSelection() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMaasBootSourceSelectionRead,
		Description: "Provides a resource to fetch a MAAS boot source selection.",

		Schema: map[string]*schema.Schema{
			"arches": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "The architecture list for this selection.",
			},
			"boot_source": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "The boot source database ID this resource is associated with.",
			},
			"labels": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "The label list for this selection.",
			},
			"os": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The operating system for this selection.",
			},
			"release": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The specific release of the operating system for this selection.",
			},
			"subarches": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Computed:    true,
				Description: "The list of subarches for this selection.",
			},
		},
	}
}

func dataSourceMaasBootSourceSelectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	bootsourceselection, err := getBootSourceSelectionByRelease(client, d.Get("boot_source").(int), d.Get("os").(string), d.Get("release").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", bootsourceselection.ID))

	tfState := map[string]interface{}{
		"arches":      bootsourceselection.Arches,
		"boot_source": bootsourceselection.BootSourceID,
		"labels":      bootsourceselection.Labels,
		"os":          bootsourceselection.OS,
		"release":     bootsourceselection.Release,
		"subarches":   bootsourceselection.Subarches,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getBootSourceSelectionByRelease(client *client.Client, boot_source int, os string, release string) (*entity.BootSourceSelection, error) {
	bootsourceselection, err := findBootSourceSelection(client, boot_source, os, release)
	if err != nil {
		return nil, err
	}
	if bootsourceselection == nil {
		return nil, fmt.Errorf("boot source selection (%s %s) was not found", os, release)
	}
	return bootsourceselection, nil
}

func findBootSourceSelection(client *client.Client, boot_source int, os string, release string) (*entity.BootSourceSelection, error) {
	bootsourceselections, err := client.BootSourceSelections.Get(boot_source)
	if err != nil {
		return nil, err
	}
	for _, d := range bootsourceselections {
		if d.OS == os && d.Release == release {
			return &d, nil
		}
	}
	return nil, nil
}
