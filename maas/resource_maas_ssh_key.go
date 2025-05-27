package maas

import (
	"context"
	"fmt"
	"strconv"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASSSHKey() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a resource to manage SSH keys in MAAS.",
		CreateContext: resourceSSHKeyCreate,
		ReadContext:   resourceSSHKeyRead,
		UpdateContext: resourceSSHKeyUpdate,
		DeleteContext: resourceSSHKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				// TODO: Implement imprt logic.
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type: 	  schema.TypeString,
				Optional: true,
				Computed: true, 
				Description: "A valid SSH public key. If specified, this key will be uploaded to MAAS. Otherwise this field will be computed.",
				ExactlyOneOf: []string{"key", "keysource"},
			},
			"keysource": {
				Type: 	schema.TypeString,
				Optional: true, 
				Computed: true,
				ExactlyOneOf: []string{"key", "keysource"},
				Description: "The source of the SSH key. Can be used to import a requesting user's SSH keys from a particular source for a specific user, specified in the format source:user. Valid sources include 'lp' for Launchpad and 'gh' for GitHub. E.g. 'lp:my_launchpad_username'",
			},
			"resource_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The URI of the SSH key resource. Computed by MAAS.",
			},
		},
	}
}

func resourceSSHKeyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	keyVal, keySpecified := d.GetOk("key")
	keysource, kesourceSpecified := d.GetOk("keysource") 

	var key *entity.SSHKey
	var err error
	switch {
		case keySpecified:
			key, err = client.SSHKeys.Create(keyVal.(string))
			if err != nil {
				return diag.FromErr(fmt.Errorf("Error creating SSH key: %v", err))
			}
		case kesourceSpecified:
			// needs to be multiple.
			key, err = client.SSHKeys.Import(keysource.(string))
			if err != nil {
				return diag.FromErr(fmt.Errorf("Error importing SSH key from source '%s': %v", keysource, err))
			}
		default:
			return diag.FromErr(fmt.Errorf("Either 'key' or 'keysource' must be specified to create an SSH key"))
	}

	d.SetId(fmt.Sprintf("%d", key.ID))

	return resourceSSHKeyRead(ctx, d, meta)
}



func resourceSSHKeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client
	sshKeyID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error getting SSH key with id: %s error: ", d.Id(), err))
	}

	sshKey, err := client.SSHKey.Get(sshKeyID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error getting SSH key with id: %s error: %v", d.Id(), err))
	}

	tfState := map[string]any{
		"key": 		sshKey.Key,
		"keysource": sshKey.Keysource,
		"resource_uri": sshKey.ResourceURI,
	}

	d.SetId(fmt.Sprintf("%d", sshKey.ID))

	err = setTerraformState(d, tfState)
	if err != nil {
		return diag.FromErr(fmt.Errorf("Error setting Terraform state for SSH key with id: %s error: %v", d.Id(), err))
	}

	return nil
}


func resourceSSHKeyUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	return nil
}
