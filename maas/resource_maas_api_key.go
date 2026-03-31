package maas

import (
	"context"
	"fmt"
	"strings"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASAPIKey() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS API keys (authorisation tokens) for the authenticated user.",
		CreateContext: resourceAPIKeyCreate,
		ReadContext:   resourceAPIKeyRead,
		UpdateContext: resourceAPIKeyUpdate,
		DeleteContext: resourceAPIKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client

				token, err := findAuthorisationTokenByKey(client, d.Id())
				if err != nil {
					return nil, err
				}

				consumerKey, tokenKey, tokenSecret, err := parseTokenString(token.Token)
				if err != nil {
					return nil, err
				}

				tfState := map[string]any{
					"id":           tokenKey,
					"name":         token.Name,
					"token_key":    tokenKey,
					"token_secret": tokenSecret,
					"consumer_key": consumerKey,
					"api_key":      token.Token,
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name (label) for the API key.",
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

func resourceAPIKeyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	name := d.Get("name").(string)

	authToken, err := client.Account.CreateAuthorisationToken(name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(authToken.TokenKey)

	apiKey := fmt.Sprintf("%s:%s:%s", authToken.ConsumerKey, authToken.TokenKey, authToken.TokenSecret)

	tfState := map[string]any{
		"name":         authToken.Name,
		"token_key":    authToken.TokenKey,
		"token_secret": authToken.TokenSecret,
		"consumer_key": authToken.ConsumerKey,
		"api_key":      apiKey,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceAPIKeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	tokenKey := d.Id()

	token, err := findAuthorisationTokenByKey(client, tokenKey)
	if err != nil {
		d.SetId("")
		return nil
	}

	consumerKey, _, tokenSecret, err := parseTokenString(token.Token)
	if err != nil {
		return diag.FromErr(err)
	}

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

func resourceAPIKeyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	if d.HasChange("name") {
		tokenKey := d.Id()
		name := d.Get("name").(string)

		if err := client.Account.UpdateTokenName(name, tokenKey); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceAPIKeyRead(ctx, d, meta)
}

func resourceAPIKeyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	if err := client.Account.DeleteAuthorisationToken(d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

// findAuthorisationTokenByKey finds an authorisation token by its token_key
// from the list of all tokens for the authenticated user.
func findAuthorisationTokenByKey(c *client.Client, tokenKey string) (*entity.AuthorisationTokenListItem, error) {
	tokens, err := c.Account.ListAuthorisationTokens()
	if err != nil {
		return nil, err
	}

	for _, t := range tokens {
		_, key, _, err := parseTokenString(t.Token)
		if err != nil {
			continue
		}

		if key == tokenKey {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("authorisation token with key %q was not found", tokenKey)
}

// findAuthorisationTokenByName finds an authorisation token by its name
// from the list of all tokens for the authenticated user.
func findAuthorisationTokenByName(c *client.Client, name string) (*entity.AuthorisationTokenListItem, error) {
	tokens, err := c.Account.ListAuthorisationTokens()
	if err != nil {
		return nil, err
	}

	for _, t := range tokens {
		if t.Name == name {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("authorisation token with name %q was not found", name)
}

// parseTokenString splits a MAAS token string (consumer_key:token_key:token_secret)
// into its three components.
func parseTokenString(token string) (consumerKey, tokenKey, tokenSecret string, err error) {
	parts := strings.SplitN(token, ":", 3)
	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid token format: expected consumer_key:token_key:token_secret, got %q", token)
	}

	return parts[0], parts[1], parts[2], nil
}
