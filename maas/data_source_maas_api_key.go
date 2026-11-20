package maas

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMAASAPIKey() *schema.Resource {
	return &schema.Resource{
		Description: "Provides details about an existing MAAS API key (authorisation token) for the authenticated user.",
		ReadContext: dataSourceAPIKeyRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name (label) of the API key to look up.",
			},
			"token_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The token key component of the API key.",
			},
			"token_secret": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The token secret component of the API key.",
			},
			"consumer_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The consumer key component of the API key.",
			},
			"api_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "The full API key in the format `consumer_key:token_key:token_secret`.",
			},
		},
	}
}

func dataSourceAPIKeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	name := d.Get("name").(string)

	token, err := findAuthorisationTokenByName(client, name)
	if err != nil {
		return diag.FromErr(err)
	}

	consumerKey, tokenKey, tokenSecret, err := parseTokenString(token.Token)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(tokenKey)

	tfState := map[string]any{
		"name":         token.Name,
		"token_key":    tokenKey,
		"token_secret": tokenSecret,
		"consumer_key": consumerKey,
		"api_key":      token.Token,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
