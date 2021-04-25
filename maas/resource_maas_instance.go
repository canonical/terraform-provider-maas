package maas

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/juju/gomaasapi"
)

func resourceMaasInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInstanceCreate,
		ReadContext:   resourceInstanceRead,
		DeleteContext: resourceInstanceDelete,

		Schema: map[string]*schema.Schema{
			"min_cpu_count": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"min_memory": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"tags": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"pool": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"distro_series": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "focal",
			},
			"hwe_kernel": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"user_data": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"cpu_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_addresses": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(gomaasapi.Controller)

	// Allocate MAAS machine
	machine, _, err := client.AllocateMachine(getMaasMachineAllocateArgs(d))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to allocate MAAS machine: %s", err))
	}

	// Save system id
	d.SetId(machine.SystemID())

	// Deploy MAAS machine
	err = deployMaasMachine(d, client, ctx)
	if err != nil {
		return diag.FromErr(err)
	}

	// Read MAAS machine info
	return resourceInstanceRead(ctx, d, m)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(gomaasapi.Controller)

	// Get MAAS machine
	machine, err := getMaasMachine(client, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get machine (%s): %s", d.Id(), err))
	}

	// Set Terraform state
	d.Set("cpu_count", machine.CPUCount())
	d.Set("memory", machine.Memory())
	d.Set("fqdn", machine.FQDN())
	d.Set("ip_addresses", machine.IPAddresses())

	return nil
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(gomaasapi.Controller)

	// Release MAAS machine
	err := client.ReleaseMachines(gomaasapi.ReleaseMachinesArgs{SystemIDs: []string{d.Id()}})
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to release machine (%s): %s", d.Id(), err))
	}

	// Wait MAAS machine to be released
	log.Printf("[DEBUG] Waiting for machine (%s) to be released\n", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Releasing"},
		Target:     []string{"Ready"},
		Refresh:    getMaasMachineStatusFunc(client, d.Id()),
		Timeout:    1 * time.Minute,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("machine (%s) couldn't be released: %s", d.Id(), err))
	}

	return nil
}

func getMaasMachineStatusFunc(client gomaasapi.Controller, systemId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		machine, err := getMaasMachine(client, systemId)
		if err != nil {
			log.Printf("[ERROR] Unable to get machine (%s) status: %s\n", systemId, err)
			return nil, "", err
		}

		log.Printf("[DEBUG] Machine (%s) status: %s\n", systemId, machine.StatusName())

		return machine, machine.StatusName(), nil
	}
}

func getMaasMachineAllocateArgs(d *schema.ResourceData) gomaasapi.AllocateMachineArgs {
	allocateMachineArgs := gomaasapi.AllocateMachineArgs{}
	if cpuCount, ok := d.GetOk("min_cpu_count"); ok {
		allocateMachineArgs.MinCPUCount = cpuCount.(int)
	}
	if memory, ok := d.GetOk("min_memory"); ok {
		allocateMachineArgs.MinMemory = memory.(int)
	}
	if tags, ok := d.GetOk("tags"); ok {
		allocateMachineArgs.Tags = convertToStringSlice(tags)
	}
	if zone, ok := d.GetOk("zone"); ok {
		allocateMachineArgs.Zone = zone.(string)
	}
	if pool, ok := d.GetOk("pool"); ok {
		allocateMachineArgs.Pool = pool.(string)
	}
	return allocateMachineArgs
}

func getMaasMachineStartArgs(d *schema.ResourceData) gomaasapi.StartArgs {
	startArgs := gomaasapi.StartArgs{}

	if userData, ok := d.GetOk("user_data"); ok {
		startArgs.UserData = base64Encode([]byte(userData.(string)))
	}
	if distroSeries, ok := d.GetOk("distro_series"); ok {
		startArgs.DistroSeries = distroSeries.(string)
	}
	if kernel, ok := d.GetOk("hwe_kernel"); ok {
		startArgs.Kernel = kernel.(string)
	}

	return startArgs
}

func deployMaasMachine(d *schema.ResourceData, client gomaasapi.Controller, ctx context.Context) error {
	// Get MAAS machine
	machine, err := getMaasMachine(client, d.Id())
	if err != nil {
		return fmt.Errorf("failed to get machine (%s): %s", d.Id(), err)
	}

	// Start MAAS machine
	err = machine.Start(getMaasMachineStartArgs(d))
	if err != nil {
		return fmt.Errorf("failed to start machine (%s): %s", machine.SystemID(), err)
	}

	// Wait for MAAS machine to be deployed
	log.Printf("[DEBUG] Waiting for machine (%s) to become deployed\n", machine.SystemID())
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Deploying"},
		Target:     []string{"Deployed"},
		Refresh:    getMaasMachineStatusFunc(client, machine.SystemID()),
		Timeout:    30 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return fmt.Errorf("machine (%s) didn't deploy: %s", machine.SystemID(), err)
	}

	return nil
}
