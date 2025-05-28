package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASSSHKeys() *schema.Resource {
	return &schema.Resource{
		Description: "Provides a resource to manage one or more SSH keys in MAAS.",
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
			"keys": {
				Type: 	  schema.TypeSet,
				Elem: 	  &schema.Schema{Type: schema.TypeString},
				Optional: true,
				Computed: true,
				ForceNew: true,
				ExactlyOneOf: []string{"keys", "keysource"},
				Description: "Valid SSH public keys. If specified, these keys will be uploaded to MAAS. Otherwise this field will be computed.",
			},
			"keysource": {
				Type: 	schema.TypeString,
				Optional: true,
				ForceNew: true,
				ExactlyOneOf: []string{"keys", "keysource"},
				Description: `The source of the SSH key. Can be used to import a requesting user's SSH key 
				from a source for a specific user, specified in the format source:user. Valid sources 
				include 'lp' for Launchpad and 'gh' for GitHub. E.g. 'lp:my_launchpad_username'. 

				Note that all keys from the source will be imported into MAAS, and all keys will be managed by this resource.`,
			},
		},
	}
}

func resourceSSHKeyCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	keySet, keySpecified := d.GetOk("keys")
	keysource, keysourceSpecified := d.GetOk("keysource") 

	var keys []entity.SSHKey
	var err error
	
	switch {
		case keySpecified:
			keyVals := convertToStringSlice(keySet.(*schema.Set).List())
			for _, key := range keyVals {
				sshKey, err := client.SSHKeys.Create(key)
				if err != nil {
					return diag.FromErr(fmt.Errorf("error creating SSH key: %v", err))
				}
				keys = append(keys, *sshKey)
			}
		case keysourceSpecified:
			// Importing from a keysource can import multiple keys, which is why a user may want to manage multiple keys together.
			keys, err = client.SSHKeys.Import(keysource.(string))
			if err != nil {
				return diag.FromErr(fmt.Errorf("error importing SSH key from source '%s': %v", keysource, err))
			}
			if len(keys) == 0 {
				return diag.FromErr(fmt.Errorf("no SSH keys imported from source '%s'", keysource))
			}
		default:
			return diag.FromErr(fmt.Errorf("either 'keys' or 'keysource' must be specified to create an SSH key"))
	}

	d.SetId(CreateIDFromKeys(keys))
	
	return resourceSSHKeyRead(ctx, d, meta)
}

// Create a SSH key state id from a list of SSH keys of the format id1/id2/id3.
func CreateIDFromKeys(keys []entity.SSHKey) string {
	sshKeyValues := make([]string, len(keys))
	for i, key := range keys {
		sshKeyValues[i] = fmt.Sprintf("%d", key.ID)
	}
	return strings.Join(sshKeyValues, "/")
}

// Split the state ID of a SSH key resource in the format "id1/id2/id3" into its component ids, where id1, id2, id3 are int Ids.
func SplitSSHKeyStateID(stateID string) ([]int, error) {
	splitID := strings.Split(stateID, "/")

	ids := make([]int, len(splitID))
	var err error
	
	// Convert each string id to an int
	for i, id := range splitID {
		ids[i], err = strconv.Atoi(id)
		if err != nil {
			return nil, err
		}
	}

	return ids, nil
}

func resourceSSHKeyRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	sshKeyIDs, err := SplitSSHKeyStateID(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error converting SSH key id to int: %v", err))
	}

	keys := make([]string, len(sshKeyIDs))
	for i, sshKeyID := range sshKeyIDs {
		sshKey, err := client.SSHKey.Get(sshKeyID)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error getting SSH key with id: %d error: %v", sshKeyID, err))
		}
		keys[i] = sshKey.Key

		if keysource, ok := d.GetOk("keysource"); ok {
			if sshKey.Keysource != keysource.(string) {
				return diag.FromErr(fmt.Errorf("SSH key with id: %d has a different keysource '%s' than the resource '%s'", sshKeyID, sshKey.Keysource, keysource.(string)))
			}
		}
	}

	tfState := map[string]any{
		"keys": 		keys,
	}

	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(fmt.Errorf("error setting Terraform state for SSH key with id: %s error: %v", d.Id(), err))
	}

	return nil
}

func resourceSSHKeyDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	sshKeyIDs, err := SplitSSHKeyStateID(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	for _, sshKeyID := range sshKeyIDs {
		if err := client.SSHKey.Delete(sshKeyID); err != nil {
			return diag.FromErr(fmt.Errorf("failed to delete SSH key with id: %d with error: %v", sshKeyID, err))
		}
	}

	return nil
}