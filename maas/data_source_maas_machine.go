package maas

import (
	"context"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func dataSourceMAASMachine() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMachineRead,

		Schema: map[string]*schema.Schema{
			"architecture": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The architecture type of the machine.",
			},
			"block_devices": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of physical block devices attached to the machine.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The ID of the block device.",
						},
						"id_path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID path of the block device.",
						},
						"model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The model of the block device.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The block device name.",
						},
						"serial": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The serial number of the block device.",
						},
						"size_gigabytes": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The size of the block device (in GB).",
						},
					},
				},
			},
			"domain": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The domain of the machine.",
			},
			"hostname": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"hostname", "pxe_mac_address"},
				Description:  "The machine hostname.",
			},
			"min_hwe_kernel": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The minimum kernel version allowed to run on this machine.",
			},
			"pool": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The resource pool of the machine.",
			},
			"power_parameters": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Serialized JSON string containing the parameters specific to the `power_type`. See [Power types](https://maas.io/docs/api#power-types) section for a list of the available power parameters for each power type.",
			},
			"power_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The power management type (e.g. `ipmi`) of the machine.",
			},
			"pxe_mac_address": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ExactlyOneOf: []string{"hostname", "pxe_mac_address"},
				Description:  "The MAC address of the machine's PXE boot NIC.",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The machine status",
			},
			"zone": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The zone of the machine.",
			},
		},
	}
}

func dataSourceMachineRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	var identifier string

	if v, ok := d.GetOk("hostname"); ok {
		identifier = v.(string)
	} else if v, ok := d.GetOk("pxe_mac_address"); ok {
		identifier = v.(string)
	}

	machine, err := getMachine(client, identifier)
	if err != nil {
		return diag.FromErr(err)
	}

	powerParams, err := client.Machine.GetPowerParameters(machine.SystemID)
	if err != nil {
		return diag.FromErr(err)
	}

	powerParamsJSON, err := structure.FlattenJsonToString(powerParams)
	if err != nil {
		return diag.FromErr(err)
	}

	allBlockDevices, err := client.BlockDevices.Get(machine.SystemID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Virtual block devices will be managed by Terraform resources, so do not add value
	// by being here. Filtering them out prevents unnecessary plan changes after
	// other resources create virtual block devices.
	physicalBlockDevices := make([]entity.BlockDevice, 0)

	for _, bd := range allBlockDevices {
		if bd.Type == "physical" {
			physicalBlockDevices = append(physicalBlockDevices, bd)
		}
	}

	tfState := map[string]any{
		"id":               machine.SystemID,
		"architecture":     machine.Architecture,
		"min_hwe_kernel":   machine.MinHWEKernel,
		"hostname":         machine.Hostname,
		"domain":           machine.Domain.Name,
		"zone":             machine.Zone.Name,
		"pool":             machine.Pool.Name,
		"power_type":       machine.PowerType,
		"power_parameters": powerParamsJSON,
		"pxe_mac_address":  machine.BootInterface.MACAddress,
		"status":           machine.StatusName,
		"block_devices":    getAllBlockDeviceMachineParameters(physicalBlockDevices),
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
