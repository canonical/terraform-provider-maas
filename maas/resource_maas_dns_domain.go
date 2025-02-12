package maas

import (
	"context"
	"fmt"
	"strconv"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMaasDnsDomain() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS DNS domains.",
		CreateContext: resourceDnsDomainCreate,
		ReadContext:   resourceDnsDomainRead,
		UpdateContext: resourceDnsDomainUpdate,
		DeleteContext: resourceDnsDomainDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client

				domain, err := getDomain(client, d.Id())
				if err != nil {
					return nil, err
				}
				tfState := map[string]interface{}{
					"id":            fmt.Sprintf("%v", domain.ID),
					"name":          domain.Name,
					"ttl":           domain.TTL,
					"authoritative": domain.Authoritative,
					"is_default":    domain.IsDefault,
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"authoritative": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Boolean value indicating if the new DNS domain is authoritative. Defaults to `false`.",
			},
			"is_default": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Boolean value indicating if the new DNS domain will be set as the default in the MAAS environment. Defaults to `false`.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the new DNS domain.",
			},
			"ttl": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The default TTL for the new DNS domain.",
			},
		},
	}
}

func resourceDnsDomainCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	domain, err := client.Domains.Create(getDomainParams(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", domain.ID))

	return resourceDnsDomainUpdate(ctx, d, meta)
}

func resourceDnsDomainRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := client.Domain.Get(id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceDnsDomainUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	domain, err := client.Domain.Update(id, getDomainParams(d))
	if err != nil {
		return diag.FromErr(err)
	}
	if d.Get("is_default").(bool) {
		if _, err := client.Domain.SetDefault(domain.ID); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDnsDomainRead(ctx, d, meta)
}

func resourceDnsDomainDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if err := client.Domain.Delete(id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getDomainParams(d *schema.ResourceData) *entity.DomainParams {
	return &entity.DomainParams{
		Name:          d.Get("name").(string),
		TTL:           d.Get("ttl").(int),
		Authoritative: d.Get("authoritative").(bool),
	}
}

func getDomain(client *client.Client, identifier string) (*entity.Domain, error) {
	domains, err := client.Domains.Get()
	if err != nil {
		return nil, err
	}
	for _, d := range domains {
		if fmt.Sprintf("%v", d.ID) == identifier || d.Name == identifier {
			return &d, nil
		}
	}
	return nil, fmt.Errorf("domain (%s) was not found", identifier)
}
