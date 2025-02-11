package maas

import (
	"context"
	"fmt"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMaasUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS users.",
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client

				user, err := getUser(client, d.Id())
				if err != nil {
					return nil, err
				}
				tfState := map[string]interface{}{
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
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	user, err := client.Users.Create(getUserParams(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(user.UserName)

	return nil
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	if _, err := client.User.Get(d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	if err := client.User.Delete(d.Id()); err != nil {
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
