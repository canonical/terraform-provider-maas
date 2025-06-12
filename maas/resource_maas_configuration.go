package maas

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceMAASConfiguration() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS configuration settings through key value pairs. See MAAS server in the [MAAS API documentation](https://maas.io/docs/api) for a full list of keys and their values. Upon destroy, the specified key will no longer be managed by Terraform, but the set value will remain set in MAAS.",
		CreateContext: resourceMAASConfigurationCreate,
		ReadContext:   resourceMAASConfigurationRead,
		UpdateContext: resourceMAASConfigurationUpdate,
		DeleteContext: resourceMAASConfigurationDelete,
		Schema: map[string]*schema.Schema{
			"key": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Key corresponding to the configuration setting.",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringDoesNotMatch(regexp.MustCompile("maas_auto_ipmi_workaround_flags"), "Key 'maas_auto_ipmi_workaround_flags' cannot currently be set through Terraform due to a [bug in MAAS](https://bugs.launchpad.net/maas/+bug/2112191).")),
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Value for the configuration setting, always specified as a string",
			},
		},
	}
}

func resourceMAASConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client
	key := d.Get("key").(string)
	value := d.Get("value").(string)

	err := client.MAASServer.Post(key, value)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error setting key %q with value %q in MAAS: %v", key, value, err))
	}

	d.SetId(key)

	return resourceMAASConfigurationRead(ctx, d, meta)
}

func resourceMAASConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	key := d.Id()

	value, err := client.MAASServer.Get(key)
	if err != nil {
		return diag.FromErr(err)
	}

	tfState := map[string]any{
		"value": NormalizeConfigValue(value),
		"key":   key,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMAASConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.HasChange("key") {
		oldVal, newVal := d.GetChange("key")
		return diag.Errorf("Changing 'key' from %v to %v is not allowed. Please recreate the resource.", oldVal, newVal)
	}

	client := meta.(*ClientConfig).Client

	err := client.MAASServer.Post(d.Get("key").(string), d.Get("value").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceMAASConfigurationRead(ctx, d, meta)
}

func resourceMAASConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Currently, settings cannot be reverted to their default values through the API. For now, just remove the key from the state.
	d.SetId("")
	return nil
}

func NormalizeConfigValue(apiValue []byte) string {
	// MAAS allows setting of some null values using empty strings i.e. "", but this will output `null` as bytes when read back for some keys e.g. remote_syslog.
	// As the user can only set null values with empty strings, we will convert `null` to an empty string instead of using something like DiffSuppressFunc.
	if bytes.Equal(apiValue, []byte("null")) {
		return ""
	}
	// Some string values are returned as quoted strings, which we will remove.
	return strings.Trim(string(apiValue), "\"")
}
