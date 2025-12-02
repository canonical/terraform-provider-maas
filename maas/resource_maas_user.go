package maas

import (
	"context"
	"fmt"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS users. *Note* You cannot use this to modify the logged in terraform user, or any users not managed by MAAS.",
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client

				user, err := getValidUser(client, d)
				if err != nil {
					return nil, err
				}

				tfState := map[string]any{
					"id":       user.UserName,
					"name":     user.UserName,
					"email":    user.Email,
					"is_admin": user.IsSuperUser,
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"email": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: isEmailAddress,
				Description:      "The user e-mail address.",
			},
			"is_admin": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Boolean value indicating if the user is a MAAS administrator. Defaults to `false`.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The user name.",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				ForceNew:    true,
				Description: "The user password.",
			},
			"transfer_to_user": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "If provided, resources owned by the deleted user will be transferred to this user.",
			},
		},
		CustomizeDiff: func(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
			name := d.Get("name").(string)
			transferTo := d.Get("transfer_to_user").(string)

			if transferTo != "" && name == transferTo {
				return fmt.Errorf("`transfer_to_user` cannot have the same value as `name` (%q). Specify another user to transfer resources to on delete", name)
			}

			return nil
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	user, err := client.Users.Create(getUserParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.UserName)

	return resourceUserRead(ctx, d, meta)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	userName := d.Id()

	user, err := client.User.Get(userName)
	if err != nil {
		return diag.FromErr(err)
	}

	tfState := map[string]interface{}{
		"email":    user.Email,
		"is_admin": user.IsSuperUser,
		"name":     user.UserName,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.Errorf("Could not set user state: %v", err)
	}

	return nil
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	deleteParams := entity.UserDeleteParams{
		UserName: d.Id(),
	}

	if user, ok := d.GetOk("transfer_to_user"); ok {
		transferUserName := user.(string)
		if transferUserName != "" {
			transferUser, err := getUser(client, transferUserName)
			if err != nil {
				return diag.Errorf("user %q to transfer resources to doesn't exist: %v", transferUserName, err)
			}

			deleteParams.TransferResourcesTo = transferUser.UserName
		}
	}

	if err := client.User.Delete(&deleteParams); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getUserParams(d *schema.ResourceData) *entity.UserParams {
	return &entity.UserParams{
		UserName:    d.Get("name").(string),
		Password:    d.Get("password").(string),
		Email:       d.Get("email").(string),
		IsSuperUser: d.Get("is_admin").(bool),
	}
}

func getUser(client *client.Client, userName string) (*entity.User, error) {
	users, err := client.Users.Get()
	if err != nil {
		return nil, err
	}

	for _, u := range users {
		if u.UserName == userName {
			return &u, nil
		}
	}

	return nil, fmt.Errorf("user (%s) was not found", userName)
}

func getValidUser(client *client.Client, d *schema.ResourceData) (*entity.User, error) {
	// ensure the user is a valid target for import
	user, err := getUser(client, d.Id())
	if err != nil {
		return nil, err
	}

	me, err := client.Users.Whoami()
	if err != nil {
		return nil, err
	}

	// terraform cannot import non-local users
	if !user.IsLocal {
		return nil, fmt.Errorf("cannot operate on non-local users, use the user service providing the user account instead")
	}

	// and we also don't allow modifying ourselves
	if user.UserName == me.UserName {
		return nil, fmt.Errorf("cannot operate on the currently logged in user %q", me.UserName)
	}

	return user, nil
}
