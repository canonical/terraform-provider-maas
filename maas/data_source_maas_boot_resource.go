package maas

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMaasBootResources() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMaasBootResourcesRead,
		Description: "Provides a data source to manage MAAS bootresources.",

		Schema: map[string]*schema.Schema{
			"boot_source": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The boot source database ID this resource is associated with.",
			},
			"boot_source_selections": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "The set of database IDs for boot source selections to attach to this boot resource",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func dataSourceMaasBootResourcesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	resources, err := getBootResources(client, "synced")
	if err != nil {
		return diag.FromErr(err)
	}
	bootsource, err := getBootSource(client)
	if err != nil {
		return diag.FromErr(err)
	}

	var selectionSet []int
	for _, res := range resources {
		parts := strings.SplitN(res.Name, "/", 2)
		if len(parts) < 2 {
			return diag.Errorf("Invalid resource name: %s", res.Name)
		}
		os, release := parts[0], parts[1]

		// avoid the bootloaders
		if strings.HasPrefix(os, "uefi") || strings.HasPrefix(os, "pxe") {
			continue
		}

		selection, err := getBootSourcesByRelease(client, bootsource.ID, os, release)
		if err != nil {
			return diag.FromErr(err)
		}
		selectionSet = append(selectionSet, selection.ID)
	}

	tfState := map[string]interface{}{
		"boot_source":            bootsource.ID,
		"boot_source_selections": unique(selectionSet),
	}

	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
