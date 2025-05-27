package maas

import (
	"context"
	"fmt"
	"strconv"

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
				client := meta.(*ClientConfig).Client
				sshKeyID, err := strconv.Atoi(d.Id())
				if err != nil {
					return nil, fmt.Errorf("error converting SSH key id to int: %v", err)
				}
				sshKey, err := client.SSHKey.Get(sshKeyID)
				if err != nil {
					return nil, fmt.Errorf("error importing SSH key with id: %s error: %v", d.Id(), err)
				}
				d.SetId(fmt.Sprintf("%d", sshKey.ID))
				return []*schema.ResourceData{d}, nil
			},
		},
		Schema: map[string]*schema.Schema{
			"key": {
				Type: 	  schema.TypeString,
				Required: true,
				ForceNew: true,
				Description: "A valid SSH public key. If specified, this key will be uploaded to MAAS. Otherwise this field will be computed.",
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

	keyVal := d.Get("key").(string)

	key, err := client.SSHKeys.Create(keyVal)
			if err != nil {
				return diag.FromErr(fmt.Errorf("error creating SSH key: %v", err))
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