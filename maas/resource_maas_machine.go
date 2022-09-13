package maas

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/maas/gomaasclient/client"
	"github.com/maas/gomaasclient/entity"
)

func resourceMaasMachine() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS machines.",
		CreateContext: resourceMachineCreate,
		ReadContext:   resourceMachineRead,
		UpdateContext: resourceMachineUpdate,
		DeleteContext: resourceMachineDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				client := m.(*client.Client)
				machine, err := getMachine(client, d.Id())
				if err != nil {
					return nil, err
				}
				powerParams, err := client.Machine.GetPowerParameters(machine.SystemID)
				if err != nil {
					return nil, err
				}
				tfState := map[string]interface{}{
					"id":               machine.SystemID,
					"power_type":       machine.PowerType,
					"power_parameters": powerParams,
					"pxe_mac_address":  machine.BootInterface.MACAddress,
					"architecture":     machine.Architecture,
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"power_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A power management type (e.g. `ipmi`).",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
					[]string{
						"amt", "apc", "dli", "eaton", "hmc", "ipmi", "manual", "moonshot",
						"mscm", "msftocs", "nova", "openbmc", "proxmox", "recs_box", "redfish",
						"sm15k", "ucsm", "vmware", "webhook", "wedge", "lxd", "virsh",
					},
					false)),
			},
			"power_parameters": {
				Type:        schema.TypeMap,
				Required:    true,
				Sensitive:   true,
				Description: "A map with the parameters specific to the `power_type`. See [Power types](https://maas.io/docs/api#power-types) section for a list of the available power parameters for each power type.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"pxe_mac_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The MAC address of the machine's PXE boot NIC.",
			},
			"architecture": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "amd64/generic",
				Description: "The architecture type of the machine. Defaults to `amd64/generic`.",
			},
			"min_hwe_kernel": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The minimum kernel version allowed to run on this machine. Only used when deploying Ubuntu. This is computed if it's not set.",
			},
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The machine hostname. This is computed if it's not set.",
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The domain of the machine. This is computed if it's not set.",
			},
			"zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The zone of the machine. This is computed if it's not set.",
			},
			"pool": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource pool of the machine. This is computed if it's not set.",
			},
		},
	}
}

func resourceMachineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Create MAAS machine
	machine, err := client.Machines.Create(getMachineParams(d), getMachinePowerParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	// Save Id
	d.SetId(machine.SystemID)

	// Wait for machine to be ready
	_, err = waitForMachineStatus(ctx, client, machine.SystemID, []string{"Commissioning", "Testing"}, []string{"Ready"})
	if err != nil {
		return diag.FromErr(err)
	}

	// Return updated machine
	return resourceMachineUpdate(ctx, d, m)
}

func resourceMachineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Get machine
	machine, err := client.Machine.Get(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Set Terraform state
	tfState := map[string]interface{}{
		"architecture":   machine.Architecture,
		"min_hwe_kernel": machine.MinHWEKernel,
		"hostname":       machine.Hostname,
		"domain":         machine.Domain.Name,
		"zone":           machine.Zone.Name,
		"pool":           machine.Pool.Name,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMachineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Update machine
	machine, err := client.Machine.Get(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if _, err := client.Machine.Update(machine.SystemID, getMachineParams(d), getMachinePowerParams(d)); err != nil {
		return diag.FromErr(err)
	}

	return resourceMachineRead(ctx, d, m)
}

func resourceMachineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Delete machine
	if err := client.Machine.Delete(d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getMachinePowerParams(d *schema.ResourceData) map[string]string {
	powerParams := d.Get("power_parameters").(map[string]interface{})
	params := make(map[string]string, len(powerParams))
	for k, v := range powerParams {
		params[fmt.Sprintf("power_parameters_%s", k)] = v.(string)
	}
	return params
}

func getMachineParams(d *schema.ResourceData) *entity.MachineParams {
	return &entity.MachineParams{
		Commission:    true,
		PowerType:     d.Get("power_type").(string),
		PXEMacAddress: d.Get("pxe_mac_address").(string),
		Architecture:  d.Get("architecture").(string),
		MinHWEKernel:  d.Get("min_hwe_kernel").(string),
		Hostname:      d.Get("hostname").(string),
		Domain:        d.Get("domain").(string),
		Zone:          d.Get("zone").(string),
		Pool:          d.Get("pool").(string),
	}
}

func getMachineStatusFunc(client *client.Client, systemId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		machine, err := client.Machine.Get(systemId)
		if err != nil {
			return nil, "", err
		}
		log.Printf("[DEBUG] Machine (%s) status: %s\n", systemId, machine.StatusName)
		return machine, machine.StatusName, nil
	}
}

func waitForMachineStatus(ctx context.Context, client *client.Client, systemID string, pendingStates []string, targetStates []string) (*entity.Machine, error) {
	log.Printf("[DEBUG] Waiting for machine (%s) status to be one of %s\n", systemID, targetStates)
	stateConf := &resource.StateChangeConf{
		Pending:    pendingStates,
		Target:     targetStates,
		Refresh:    getMachineStatusFunc(client, systemID),
		Timeout:    30 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	result, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}
	return result.(*entity.Machine), nil
}

func getMachine(client *client.Client, identifier string) (*entity.Machine, error) {
	machines, err := client.Machines.Get()
	if err != nil {
		return nil, err
	}
	for _, m := range machines {
		if m.SystemID == identifier || m.Hostname == identifier || m.FQDN == identifier {
			return &m, nil
		}
	}
	return nil, fmt.Errorf("machine (%s) not found", identifier)
}
