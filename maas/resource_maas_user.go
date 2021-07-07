package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

func resourceMaasUser() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				client := m.(*client.Client)
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
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
				ForceNew:  true,
			},
			"email": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: isEmailAddress,
			},
			"is_admin": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
		},
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	user, err := client.Users.Create(getUserParams(d))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(user.UserName)

	return nil
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	if _, err := client.User.Get(d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

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
