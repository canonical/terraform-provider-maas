package maas

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ionutbalutoiu/gomaasclient/api/endpoint"
	"github.com/ionutbalutoiu/gomaasclient/gmaw"
	"github.com/ionutbalutoiu/gomaasclient/maas"
	"github.com/juju/gomaasapi"
)

func resourceMaasMachine() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMachineCreate,
		ReadContext:   resourceMachineRead,
		UpdateContext: resourceMachineUpdate,
		DeleteContext: resourceMachineDelete,

		Schema: map[string]*schema.Schema{
			"power_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"power_parameters": {
				Type:     schema.TypeMap,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"pxe_mac_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"architecture": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "amd64",
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
	client := m.(*gomaasapi.MAASObject)

	// Create MAAS machine
	machineParams, powerParams, err := getMachineCreateParams(d)
	if err != nil {
		return diag.FromErr(err)
	}
	machinesManager := maas.NewMachinesManager(gmaw.NewMachines(client))
	machine, err := machinesManager.Create(machineParams, powerParams)
	if err != nil {
		return diag.FromErr(err)
	}

	// Save Id
	d.SetId(machine.SystemID)

	// Wait for machine to be ready
	log.Printf("[DEBUG] Waiting for machine (%s) to become ready\n", machine.SystemID)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Commissioning", "Testing"},
		Target:     []string{"Ready"},
		Refresh:    getMachineStatusFunc(client, machine.SystemID),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("machine (%s) didn't become ready within allowed timeout: %s", machine.SystemID, err))
	}

	// Return updated machine
	return resourceMachineUpdate(ctx, d, m)
}

func resourceMachineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*gomaasapi.MAASObject)

	// Get machine
	machineManager, err := maas.NewMachineManager(d.Id(), gmaw.NewMachine(client))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get machine (%s) manager: %s", d.Id(), err))
	}
	machine := machineManager.Current()

	// Set Terraform state
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
	client := m.(*gomaasapi.MAASObject)

	// Update machine
	machineManager, err := maas.NewMachineManager(d.Id(), gmaw.NewMachine(client))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get machine (%s) manager: %s", d.Id(), err))
	}
	err = machineManager.Update(getMachineUpdateParams(d, machineManager.Current()))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceMachineRead(ctx, d, m)
}

func resourceMachineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*gomaasapi.MAASObject)

	// Delete machine
	machineManager, err := maas.NewMachineManager(d.Id(), gmaw.NewMachine(client))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get machine (%s) manager: %s", d.Id(), err))
	}
	err = machineManager.Delete()
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getMachineCreateParams(d *schema.ResourceData) (*endpoint.MachineParams, interface{}, error) {
	params := endpoint.MachineParams{
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
	powerParams := d.Get("power_parameters").(map[string]interface{})

	if params.PowerType == "virsh" {
		virshParams := endpoint.VirshPowerParams{}
		if p, ok := powerParams["power_parameters_power_address"]; ok {
			virshParams.PowerAddress = p.(string)
		}
		if p, ok := powerParams["power_parameters_power_password"]; ok {
			virshParams.PowerPassword = p.(string)
		}
		if p, ok := powerParams["power_parameters_power_id"]; ok {
			virshParams.PowerID = p.(string)
		}
		return &params, virshParams, nil
	}

	return nil, nil, fmt.Errorf("machine power type %s is not supported", params.PowerType)
}

func getMachineUpdateParams(d *schema.ResourceData, machine *endpoint.Machine) *endpoint.MachineParams {
	params := endpoint.MachineParams{
		CPUCount:     machine.CPUCount,
		Memory:       machine.Memory,
		SwapSize:     machine.SwapSize,
		Architecture: machine.Architecture,
		MinHWEKernel: machine.MinHWEKernel,
		PowerType:    machine.PowerType,
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
