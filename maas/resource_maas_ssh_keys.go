package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASSSHKeys() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage one or many SSH keys in MAAS.",
		CreateContext: resourceSSHKeyCreate,
		ReadContext:   resourceSSHKeyRead,
		DeleteContext: resourceSSHKeyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"keys": {
				Type:         schema.TypeSet,
				Elem:         &schema.Schema{Type: schema.TypeString},
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"keys", "keysource"},
				Description:  "Valid SSH public keys. If specified, these keys will be uploaded to MAAS. Otherwise this field will be computed.",
			},
			"keysource": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ExactlyOneOf: []string{"keys", "keysource"},
				Description:  "The source of the SSH key(s). Can be used to 'import' a requesting user's SSH keys from a source for a specific user into MAAS, specified in the format source:user. Valid sources include 'lp' for Launchpad and 'gh' for GitHub. E.g. 'lp:my_launchpad_username'. **Note** all keys from the source will be imported into MAAS, and all keys will be managed by this resource. Keysources are not supported for import, so the expected keys must be specified using the `keys` field.",
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
		keys, err = createSSHKeysFromKeySet(keySet.(*schema.Set), client)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error creating SSH keys from key set: %v", err))
		}
	case keysourceSpecified:
		keys, err = importSSHKeysFromKeysource(keysource.(string), client)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error importing SSH keys from keysource: %v", err))
		}
	default:
		return diag.FromErr(fmt.Errorf("either 'keys' or 'keysource' must be specified to create an SSH key"))
	}

	d.SetId(CreateIDFromKeys(keys))

	return resourceSSHKeyRead(ctx, d, meta)
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
		"keys": keys,
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

// Create SSH keys in MAAS from a set of keys as strings.
func createSSHKeysFromKeySet(keySet *schema.Set, client *client.Client) ([]entity.SSHKey, error) {
	keyVals := convertToStringSlice(keySet.List())
	keys := make([]entity.SSHKey, len(keyVals))

	for i, key := range keyVals {
		sshKey, err := client.SSHKeys.Create(key)
		if err != nil {
			return nil, fmt.Errorf("error creating SSH key: %v", err)
		}

		keys[i] = *sshKey
	}

	return keys, nil
}

// importSSHKeysFromKeysource 'imports' SSH keys from a keysource, e.g. Launchpad or GitHub, into MAAS. This can import multiple keys.
func importSSHKeysFromKeysource(keysource string, client *client.Client) ([]entity.SSHKey, error) {
	keys, err := client.SSHKeys.Import(keysource)
	if err != nil {
		return nil, fmt.Errorf("error importing SSH key from source '%s': %v", keysource, err)
	}

	if len(keys) == 0 {
		return nil, fmt.Errorf("no SSH keys imported from source '%s'", keysource)
	}

	return keys, nil
}

// CreateIDFromKeys creates a SSH key state id from a list of SSH keys of the format id1/id2/id3.
func CreateIDFromKeys(keys []entity.SSHKey) string {
	sshKeyValues := make([]string, len(keys))
	for i, key := range keys {
		sshKeyValues[i] = fmt.Sprintf("%d", key.ID)
	}

	return strings.Join(sshKeyValues, "/")
}

// SplitSSHKeyStateID splits the state ID of a SSH key resource in the format "id1/id2/id3" into its component ids, where id1, id2, id3 are int ids.
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
