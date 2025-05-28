package maas

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASPackageRepositories() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS package repositories.",
		CreateContext: resourcePackageRepositoriesCreate,
		ReadContext:   resourcePackageRepositoriesRead,
		UpdateContext: resourcePackageRepositoriesUpdate,
		DeleteContext: resourcePackageRepositoriesDelete,

		Schema: map[string]*schema.Schema{
			"arches": {
				Type:        schema.TypeSet,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "The list of supported architectures.",
			},
			"components": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Description:   "The list of components to enable. Only applicable to custom repositories.",
				ConflictsWith: []string{"disabled_components", "disabled_pockets"},
			},
			"disable_sources": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Disable deb-src lines.",
			},
			"disabled_components": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Description:   "The list of components to disable. Only applicable to the default Ubuntu repositories.",
				ConflictsWith: []string{"components", "distributions"},
			},
			"disabled_pockets": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Description:   "The list of pockets to disable.",
				ConflictsWith: []string{"components", "distributions"},
			},
			"distributions": {
				Type:          schema.TypeSet,
				Optional:      true,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Description:   "Which package distributions to include.",
				ConflictsWith: []string{"disabled_components", "disabled_pockets"},
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Whether or not the repository is enabled.",
			},
			"key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "The authentication key to use with the repository.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the package repository.",
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The url of the package repository.",
			},
		},
	}
}

func resourcePackageRepositoriesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	params := &entity.PackageRepositoryParams{
		Name:               d.Get("name").(string),
		URL:                d.Get("url").(string),
		Distributions:      listAsString(d.Get("distributions").(*schema.Set).List()),
		DisabledPockets:    listAsString(d.Get("disabled_pockets").(*schema.Set).List()),
		DisabledComponents: listAsString(d.Get("disabled_components").(*schema.Set).List()),
		Components:         listAsString(d.Get("components").(*schema.Set).List()),
		Arches:             listAsString(d.Get("arches").(*schema.Set).List()),
		Key:                d.Get("key").(string),
		DisableSources:     d.Get("disable_sources").(bool),
		Enabled:            d.Get("enabled").(bool),
	}

	repo, err := client.PackageRepositories.Create(params)
	if err != nil {
		return diag.Errorf("Could not create Package Repository: %v", err)
	}

	d.SetId(fmt.Sprintf("%v", repo.ID))

	return resourcePackageRepositoriesRead(ctx, d, meta)
}

func resourcePackageRepositoriesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	repo, err := client.PackageRepository.Get(id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", repo.ID))

	// Update the Terraform state
	tfstate := map[string]interface{}{
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

	return nil
}

func resourcePackageRepositoriesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	params := &entity.PackageRepositoryParams{
		Name:               d.Get("name").(string),
		URL:                d.Get("url").(string),
		Distributions:      d.Get("distributions").(string),
		DisabledPockets:    listAsString(d.Get("disabled_pockets").(*schema.Set).List()),
		DisabledComponents: listAsString(d.Get("disabled_components").(*schema.Set).List()),
		Components:         listAsString(d.Get("components").(*schema.Set).List()),
		Arches:             listAsString(d.Get("arches").(*schema.Set).List()),
		Key:                d.Get("key").(string),
		DisableSources:     d.Get("disable_sources").(bool),
		Enabled:            d.Get("enabled").(bool),
	}

	if _, err := client.PackageRepository.Update(id, params); err != nil {
		return diag.FromErr(err)
	}

	return resourcePackageRepositoriesRead(ctx, d, meta)
}

func resourcePackageRepositoriesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	err = client.PackageRepository.Delete(id)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func listAsString(stringList []interface{}) string {
	if len(stringList) == 0 {
		return ""
	}
	asList, _ := json.Marshal(stringList)
	return string(asList)
}
