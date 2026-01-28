package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMAASPackageRepository() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMAASPackageRepositoryRead,
		Description: "Provides a data source to fetch MAAS package repositories.",

		Schema: map[string]*schema.Schema{
			"arches": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of supported architectures.",
			},
			"components": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of components that are enabled. Only applicable to custom repositories.",
			},
			"disable_sources": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether deb-src lines are disabled for this repository.",
			},
			"disabled_components": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of components that are disabled. Only applicable to the default Ubuntu repositories.",
			},
			"disabled_pockets": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of pockets that are disabled.",
			},
			"distributions": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of included package distributions.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether or not the repository is enabled.",
			},
			"key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The authentication key used with the repository.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the package repository.\n*Note*: the Name field for the Ubuntu Archive and Ubuntu Ports repos are `main_archive` and `ports_archive` respectively. As they are default resources, the MAAS UI shows a different name to their internal database name entry.",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URL of the package repository.",
			},
		},
	}
}

func dataSourceMAASPackageRepositoryRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	repo, err := getRepo(client, d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	tfstate := map[string]any{
		"arches":              repo.Arches,
		"components":          repo.Components,
		"disable_sources":     repo.DisableSources,
		"disabled_components": repo.DisabledComponents,
		"disabled_pockets":    repo.DisabledPockets,
		"distributions":       repo.Distributions,
		"enabled":             repo.Enabled,
		"key":                 repo.Key,
		"name":                repo.Name,
		"url":                 repo.URL,
	}

	if err := setTerraformState(d, tfstate); err != nil {
		return diag.Errorf("Could not set Package Repository state: %v", err)
	}

	d.SetId(fmt.Sprintf("%v", repo.ID))

	return nil
}
