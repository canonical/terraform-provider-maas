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

func resourceMaasVMHostMachine() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVMHostMachineCreate,
		ReadContext:   resourceVMHostMachineRead,
		UpdateContext: resourceVMHostMachineUpdate,
		DeleteContext: resourceVMHostMachineDelete,

		Schema: map[string]*schema.Schema{
			"vm_host": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cores": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"pinned_cores": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"storage": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"interfaces": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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

func resourceVMHostMachineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Find VM host
	vmHost, err := findVMHost(client, d.Get("vm_host").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// Create VM host machine
	params := getVMHostMachineCreateParams(d)
	machine, err := client.Pod.Compose(vmHost.ID, params)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set Terraform state
	if err := d.Set("cores", params.Cores); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pinned_cores", params.PinnedCores); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("memory", params.Memory); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("storage", params.Storage); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("interfaces", params.Interfaces); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(machine.SystemID)

	// Wait for VM host machine to be ready
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
		return diag.FromErr(err)
	}

	// Return updated VM host machine
	return resourceVMHostMachineUpdate(ctx, d, m)
}

func resourceVMHostMachineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Get VM host machine
	machine, err := client.Machine.Get(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Set Terraform state
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

func resourceVMHostMachineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Update VM host machine
	machine, err := client.Machine.Get(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.Machine.Update(machine.SystemID, getVMHostMachineUpdateParams(d, machine), map[string]string{})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceVMHostMachineRead(ctx, d, m)
}

func resourceVMHostMachineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Delete VM host machine
	err := client.Machine.Delete(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func findVMHost(client *client.Client, vmHostIdentifier string) (*entity.Pod, error) {
	vmHosts, err := client.Pods.Get()
	if err != nil {
		return nil, err
	}

	for _, vmHost := range vmHosts {
		if fmt.Sprintf("%v", vmHost.ID) == vmHostIdentifier || vmHost.Name == vmHostIdentifier {
			return &vmHost, err
		}
	}

	return nil, fmt.Errorf("VM host (%s) not found", vmHostIdentifier)
}

func getVMHostMachineCreateParams(d *schema.ResourceData) *entity.PodMachineParams {
	params := entity.PodMachineParams{}

	if p, ok := d.GetOk("cores"); ok {
		params.Cores = p.(int)
	}
	if p, ok := d.GetOk("pinned_cores"); ok {
		params.PinnedCores = p.(int)
	}
	if p, ok := d.GetOk("memory"); ok {
		params.Memory = p.(int)
	}
	if p, ok := d.GetOk("storage"); ok {
		params.Storage = p.(string)
	}
	if p, ok := d.GetOk("interfaces"); ok {
		params.Interfaces = p.(string)
	}
	if p, ok := d.GetOk("hostname"); ok {
		params.Hostname = p.(string)
	}

	return &params
}

func getVMHostMachineUpdateParams(d *schema.ResourceData, machine *entity.Machine) *entity.MachineParams {
	params := entity.MachineParams{
		CPUCount:     machine.CPUCount,
		Memory:       machine.Memory,
		SwapSize:     machine.SwapSize,
		Architecture: machine.Architecture,
		MinHWEKernel: machine.MinHWEKernel,
		PowerType:    machine.PowerType,
		Description:  machine.Description,
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
