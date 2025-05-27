package maas

import (
	"context"
	"fmt"
	"log"
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
				ForceNew: true,
				Description: "A valid SSH public key. If specified, this key will be uploaded to MAAS. Otherwise this field will be computed.",
				ExactlyOneOf: []string{"key", "keysource"},
			},
			"keysource": {
				Type: 	schema.TypeString,
				Optional: true, 
				Computed: true,
				ForceNew: true,
				ExactlyOneOf: []string{"key", "keysource"},
				Description: `The source of the SSH key. Can be used to import a requesting user's SSH key 
				from a source for a specific user, specified in the format source:user. Valid sources 
				include 'lp' for Launchpad and 'gh' for GitHub. E.g. 'lp:my_launchpad_username'. 
				Note that if more than one key is obtained, the first key will be imported.`,
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
	keysource, keysourceSpecified := d.GetOk("keysource") 

	var key *entity.SSHKey
	var err error
	switch {
		case keySpecified:
			key, err = client.SSHKeys.Create(keyVal.(string))
			if err != nil {
				return diag.FromErr(fmt.Errorf("error creating SSH key: %v", err))
			}
		case keysourceSpecified:
			// needs to be multiple.
			keys, err := client.SSHKeys.Import(keysource.(string))
			if err != nil {
				return diag.FromErr(fmt.Errorf("error importing SSH key from source '%s': %v", keysource, err))
			}
			if len(keys) == 0 {
				return diag.FromErr(fmt.Errorf("no SSH keys imported from source '%s'", keysource))
			}
			if len(keys) > 1 {
				log.Printf("[WARN] Multiple SSH keys found for source '%s'. Using first key.", keysource)
			}
			key = &keys[0]
		default:
			return diag.FromErr(fmt.Errorf("either 'key' or 'keysource' must be specified to create an SSH key"))
	}

	d.SetId(fmt.Sprintf("%d", key.ID))

	return resourceSSHKeyRead(ctx, d, meta)
}



func resourceSSHKeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client
	sshKeyID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error converting SSH key id to int: %v", err))
	}

	sshKey, err := client.SSHKey.Get(sshKeyID)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting SSH key with id: %s error: %v", d.Id(), err))
	}

	tfState := map[string]any{
		"key": 		sshKey.Key,
		"keysource": sshKey.Keysource,
		"resource_uri": sshKey.ResourceURI,
	}

	d.SetId(fmt.Sprintf("%d", sshKey.ID))

	err = setTerraformState(d, tfState)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error setting Terraform state for SSH key with id: %s error: %v", d.Id(), err))
	}

	return nil
}

func resourceSSHKeyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	sshKeyID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	if err := client.SSHKey.Delete(sshKeyID); err != nil {
		return diag.FromErr(fmt.Errorf("failed to delete SSH key with id: %d with error: %v", sshKeyID, err))
	}

	return nil
}