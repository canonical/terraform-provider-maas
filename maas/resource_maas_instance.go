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
			"distro_series": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "focal",
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
		return diag.FromErr(fmt.Errorf("error allocating MAAS machine: %s", err))
	}

	// Save system id
	d.SetId(machine.SystemID())

	// Deploy MAAS machine
	machine.Start(getMaasMachineStartArgs(d))
	log.Printf("[DEBUG] Waiting for machine (%s) to become active\n", machine.SystemID())
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Deploying"},
		Target:     []string{"Deployed"},
		Refresh:    getMaasMachineStatusFunc(client, machine.SystemID()),
		Timeout:    30 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	if _, err = stateConf.WaitForState(); err != nil {
		releaseArgs := gomaasapi.ReleaseMachinesArgs{SystemIDs: []string{machine.SystemID()}}
		if err = client.ReleaseMachines(releaseArgs); err != nil {
			log.Printf("[WARN] Unable to release machine (%s)\n", machine.SystemID())
		}
		return diag.FromErr(fmt.Errorf("error waiting for machine (%s) to become deployed: %s", machine.SystemID(), err))
	}

	// Read MAAS machine info
	return resourceInstanceRead(ctx, d, m)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(gomaasapi.Controller)

	// Get MAAS machine
	machine, err := getMaasMachine(client, d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting machine (%s): %s", d.Id(), err))
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
		return diag.FromErr(fmt.Errorf("error releasing machine with system id %s: %s", d.Id(), err))
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
	return allocateMachineArgs
}

func getMaasMachineStartArgs(d *schema.ResourceData) gomaasapi.StartArgs {
	startArgs := gomaasapi.StartArgs{}
	if distroSeries, ok := d.GetOk("distro_series"); ok {
		startArgs.DistroSeries = distroSeries.(string)
	}
	return startArgs
}
