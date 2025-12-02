package maas

import (
	"context"
	"fmt"
	"strconv"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceMAASPackageRepositories() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS package repositories.\n*Note*: The two Ubuntu archives that ship with MAAS are import-only Terraform resources, only custom repositories can be created or destroyed.",
		CreateContext: resourcePackageRepositoriesCreate,
		ReadContext:   resourcePackageRepositoriesRead,
		UpdateContext: resourcePackageRepositoriesUpdate,
		DeleteContext: resourcePackageRepositoriesDelete,

		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client

				repo, err := getRepo(client, d.Id())
				if err != nil {
					return nil, err
				}

				d.SetId(fmt.Sprintf("%v", repo.ID))

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"arches": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice(
						[]string{"amd64", "arm64", "armhf", "i386", "ppc64el", "s390x"},
						false,
					),
				},
				Description: "The list of supported architectures.",
			},
			"components": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description:   "The list of components to enable. Only applicable to custom repositories.",
				ConflictsWith: []string{"disabled_components"},
			},
			"disable_sources": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Disable deb-src lines.",
			},
			"disabled_components": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice(
						[]string{"restricted", "universe", "multiverse"},
						false,
					),
				},
				Description:   "The list of components to disable. Only applicable to the default Ubuntu repositories.",
				ConflictsWith: []string{"components"},
			},
			"disabled_pockets": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice(
						[]string{"updates", "security", "backports"},
						false,
					),
				},
				Description: "The list of pockets to disable. This only applies to Ubuntu repositories; custom or not.",
			},
			"distributions": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Which package distributions to include.",
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
				Description: "The name of the package repository.\n*Note*: the Name field for the Ubuntu Archive and Ubuntu Ports repos are `main_archive` and `ports_archive` respectively. As they are default resources, the MAAS UI shows a different name to their internal database name entry.",
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URL of the package repository.",
			},
		},
	}
}

func resourcePackageRepositoriesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	disabledComponents := d.Get("disabled_components").(*schema.Set).List()

	if len(disabledComponents) > 0 {
		return diag.Errorf("`disabled_components` are used for Ubuntu repos, which cannot be created, only imported. Specify `components` for custom repos instead.")
	}

	params := &entity.PackageRepositoryParams{
		Name:               d.Get("name").(string),
		URL:                d.Get("url").(string),
		Distributions:      listAsString(d.Get("distributions").(*schema.Set).List()),
		DisabledPockets:    listAsString(d.Get("disabled_pockets").(*schema.Set).List()),
		DisabledComponents: listAsString(disabledComponents),
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
		Distributions:      listAsString(d.Get("distributions").(*schema.Set).List()),
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

func getRepo(client *client.Client, identifier string) (*entity.PackageRepository, error) {
	repos, err := client.PackageRepositories.Get()
	if err != nil {
		return nil, err
	}

	for _, repo := range repos {
		if repo.URL == identifier || repo.Name == identifier || fmt.Sprintf("%d", repo.ID) == identifier {
			return &repo, nil
		}
	}

	return nil, fmt.Errorf("could not find repo with identifier %q", identifier)
}
