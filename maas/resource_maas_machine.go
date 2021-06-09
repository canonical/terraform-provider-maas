package maas

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

func resourceMaasMachine() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMachineCreate,
		ReadContext:   resourceMachineRead,
		UpdateContext: resourceMachineUpdate,
		DeleteContext: resourceMachineDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				client := m.(*client.Client)
				machine, err := findMachine(client, d.Id())
				if err != nil {
					return nil, err
				}
				d.SetId(machine.SystemID)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"power_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"power_parameters": {
				Type:      schema.TypeMap,
				Required:  true,
				Sensitive: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"pxe_mac_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"architecture": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "amd64/generic",
			},
			"min_hwe_kernel": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"domain": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"pool": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceMachineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Create MAAS machine
	machine, err := client.Machines.Create(getMachineCreateParams(d), getMachinePowerParams(d))
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
	if err := d.Set("power_type", machine.PowerType); err != nil {
		return diag.FromErr(err)
	}
	powerParams, err := client.Machine.GetPowerParameters(machine.SystemID)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("power_parameters", powerParams); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pxe_mac_address", machine.BootInterface.MACAddress); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("architecture", machine.Architecture); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("min_hwe_kernel", machine.MinHWEKernel); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("hostname", machine.Hostname); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("domain", machine.Domain.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("zone", machine.Zone.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pool", machine.Pool.Name); err != nil {
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
	_, err = client.Machine.Update(machine.SystemID, getMachineUpdateParams(d, machine), getMachinePowerParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceMachineRead(ctx, d, m)
}

func resourceMachineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Delete machine
	err := client.Machine.Delete(d.Id())
	if err != nil {
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

func getMachineCreateParams(d *schema.ResourceData) *entity.MachineParams {
	params := entity.MachineParams{
		PowerType:     d.Get("power_type").(string),
		PXEMacAddress: d.Get("pxe_mac_address").(string),
		Commission:    true,
	}

	if p, ok := d.GetOk("architecture"); ok {
		params.Architecture = p.(string)
	}
	if p, ok := d.GetOk("min_hwe_kernel"); ok {
		params.MinHWEKernel = p.(string)
	}
	if p, ok := d.GetOk("hostname"); ok {
		params.Hostname = p.(string)
	}
	if p, ok := d.GetOk("domain"); ok {
		params.Domain = p.(string)
	}

	return &params
}

func getMachineUpdateParams(d *schema.ResourceData, machine *entity.Machine) *entity.MachineParams {
	params := entity.MachineParams{
		PowerType:    d.Get("power_type").(string),
		CPUCount:     machine.CPUCount,
		Memory:       machine.Memory,
		SwapSize:     machine.SwapSize,
		Architecture: machine.Architecture,
		MinHWEKernel: machine.MinHWEKernel,
		Description:  machine.Description,
	}

	if p, ok := d.GetOk("architecture"); ok {
		params.Architecture = p.(string)
	}
	if p, ok := d.GetOk("min_hwe_kernel"); ok {
		params.MinHWEKernel = p.(string)
	}
	if p, ok := d.GetOk("hostname"); ok {
		params.Hostname = p.(string)
	}
	if p, ok := d.GetOk("domain"); ok {
		params.Domain = p.(string)
	}
	if p, ok := d.GetOk("zone"); ok {
		params.Zone = p.(string)
	}
	if p, ok := d.GetOk("pool"); ok {
		params.Pool = p.(string)
	}

	return &params
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

func findMachine(client *client.Client, identifier string) (*entity.Machine, error) {
	machines, err := client.Machines.Get()
	if err != nil {
		return nil, err
	}

	for _, m := range machines {
		if m.SystemID == identifier || m.Hostname == identifier || m.FQDN == identifier {
			return &m, nil
		}
	}

	return nil, fmt.Errorf("machine '%s' not found", identifier)
}
