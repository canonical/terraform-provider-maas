package maas

import (
	"context"
	"fmt"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const defaultURL = "http://images.maas.io/ephemeral-v3/stable/"
const snapKeyring = "/snap/maas/current/usr/share/keyrings/ubuntu-cloudimage-keyring.gpg"
const debKeyring = "/usr/share/keyrings/ubuntu-cloudimage-keyring.gpg"

func resourceMAASBootSource() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage the MAAS boot source.",
		CreateContext: resourceBootSourceCreate,
		ReadContext:   resourceBootSourceRead,
		UpdateContext: resourceBootSourceUpdate,
		DeleteContext: resourceBootSourceDelete,

		Schema: map[string]*schema.Schema{
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The creation time of the boot source.",
			},
			"keyring_data": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The data on the keyring for the boot source.",
				ExactlyOneOf: []string{"keyring_filename"},
			},
			"keyring_filename": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "The filename on the keyring for the boot source.",
				ExactlyOneOf: []string{"keyring_data"},
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The time of most recent update of the boot source.",
			},
			"url": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The URL of the boot source.",
			},
		},
	}
}

func resourceBootSourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	bootsource, err := getBootSource(client)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", bootsource.ID))

	// We create by transmuting the single boot source
	bootsourceParams := entity.BootSourceParams{
		KeyringData:     d.Get("keyring_data").(string),
		KeyringFilename: d.Get("keyring_filename").(string),
		URL:             d.Get("url").(string),
	}

	if _, err := client.BootSource.Update(bootsource.ID, &bootsourceParams); err != nil {
		return diag.FromErr(err)
	}

	return resourceBootSourceRead(ctx, d, meta)
}

func resourceBootSourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	bootsource, err := getBootSource(client)
	if err != nil {
		return diag.FromErr(err)
	}

	tfState := map[string]interface{}{
		"created":          bootsource.Created,
		"keyring_data":     bootsource.KeyringData,
		"keyring_filename": bootsource.KeyringFilename,
		"updated":          bootsource.Updated,
		"url":              bootsource.URL,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBootSourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	bootsource, err := getBootSource(client)
	if err != nil {
		return diag.FromErr(err)
	}

	bootsourceParams := entity.BootSourceParams{
		KeyringData:     d.Get("keyring_data").(string),
		KeyringFilename: d.Get("keyring_filename").(string),
		URL:             d.Get("url").(string),
	}

	if _, err := client.BootSource.Update(bootsource.ID, &bootsourceParams); err != nil {
		return diag.FromErr(err)
	}

	return resourceBootSourceRead(ctx, d, meta)
}

func resourceBootSourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	clientConfig := meta.(*ClientConfig)

	bootsource, err := getBootSource(clientConfig.Client)
	if err != nil {
		return diag.FromErr(err)
	}

	var keyring string
	if clientConfig.InstallationMethod == "snap" {
		keyring = snapKeyring
	} else {
		keyring = debKeyring
	}

	// We delete by changing the url and keyring fields to the default values
	bootsourceParams := entity.BootSourceParams{
		URL:             defaultURL,
		KeyringData:     "",
		KeyringFilename: keyring,
	}

	if _, err := clientConfig.Client.BootSource.Update(bootsource.ID, &bootsourceParams); err != nil {
		return diag.FromErr(err)
	}

	return resourceBootSourceRead(ctx, d, meta)
}

func getBootSource(client *client.Client) (*entity.BootSource, error) {
	bootsources, err := client.BootSources.Get()
	if err != nil {
		return nil, err
	}
	if len(bootsources) == 0 {
		return nil, fmt.Errorf("boot source was not found")
	}
	if len(bootsources) > 1 {
		return nil, fmt.Errorf("expected a single boot source")
	}
	return &bootsources[0], nil
}
