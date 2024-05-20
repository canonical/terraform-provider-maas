package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/maas/gomaasclient/client"
)

func dataSourceMaasVMHost() *schema.Resource {
	return &schema.Resource{
		Description: "Provides details about an existing MAAS VM host.",
		ReadContext: dataSourceVMHostRead,

		Schema: map[string]*schema.Schema{
			"certificate": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Certificate to use for power control of the LXD VM host.",
			},
			"cpu_over_commit_ratio": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The VM host CPU overcommit ratio.",
			},
			"default_macvlan_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The VM host default macvlan mode. Supported values are: `bridge`, `passthru`, `private`, `vepa`.",
			},
			"key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key to use for power control of the LXD VM host.",
			},
			"memory_over_commit_ratio": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "The VM host RAM memory overcommit ratio.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The VM host name.",
			},
			"pool": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The VM host pool name.",
			},
			"power_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Address that gives MAAS access to the VM host power control.",
			},
			"power_pass": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "User password to use for power control of the VM host.",
			},
			"power_user": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "User name to use for power control of the VM host.",
			},
			"resources_cores_total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The VM host total number of CPU cores.",
			},
			"resources_local_storage_total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The VM host total local storage (in bytes).",
			},
			"resources_memory_total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The VM host total RAM memory (in MB).",
			},
			"tags": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of VM host tag names.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The VM host type. Supported values are: `lxd`, `virsh`.",
			},
			"zone": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The VM host zone name.",
			},
		},
	}
}

func dataSourceVMHostRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*client.Client)

	// Get VM host details
	vmHost, err := getVMHost(client, d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", vmHost.ID))

	vmHostParameters, err := client.VMHost.GetParameters(vmHost.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set Terraform state
	tfState := map[string]interface{}{
		"cpu_over_commit_ratio":         vmHost.CPUOverCommitRatio,
		"default_macvlan_mode":          vmHost.DefaultMACVLANMode,
		"memory_over_commit_ratio":      vmHost.MemoryOverCommitRatio,
		"name":                          vmHost.Name,
		"pool":                          vmHost.Pool.Name,
		"resources_cores_total":         vmHost.Total.Cores,
		"resources_local_storage_total": vmHost.Total.LocalStorage,
		"resources_memory_total":        vmHost.Total.Memory,
		"zone":                          vmHost.Zone.Name,
		"type":                          vmHost.Type,
		"tags":                          vmHost.Tags,
	}

	if powerAddress, ok := vmHostParameters["power_address"]; ok {
		tfState["power_address"] = powerAddress
	}
	if powerAddress, ok := vmHostParameters["password"]; ok {
		tfState["power_pass"] = powerAddress
	}
	if powerAddress, ok := vmHostParameters["power_user"]; ok {
		tfState["power_user"] = powerAddress
	}
	if powerAddress, ok := vmHostParameters["certificate"]; ok {
		tfState["certificate"] = powerAddress
	}
	if powerAddress, ok := vmHostParameters["key"]; ok {
		tfState["key"] = powerAddress
	}

	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
