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
	client := m.(*gomaasapi.MAASObject)

	// Allocate MAAS machine
	machinesManager := maas.NewMachinesManager(gmaw.NewMachines(client))
	machine, err := machinesManager.Allocate(getMachinesAllocateParams(d))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to allocate MAAS machine: %s", err))
	}

	// Save system id
	d.SetId(machine.SystemID)

	// Deploy MAAS machine
	machineManager, err := maas.NewMachineManager(machine.SystemID, gmaw.NewMachine(client))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get machine (%s) manager: %s", machine.SystemID, err))
	}
	err = machineManager.Deploy(getMachineDeployParams(d))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to deploy machine (%s): %s", machine.SystemID, err))
	}

	// Wait for MAAS machine to be deployed
	log.Printf("[DEBUG] Waiting for machine (%s) to become deployed\n", machine.SystemID)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Deploying"},
		Target:     []string{"Deployed"},
		Refresh:    getMachineStatusFunc(client, machine.SystemID),
		Timeout:    30 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("machine (%s) didn't deploy within allowed timeout: %s", machine.SystemID, err))
	}

	// Read MAAS machine info
	return resourceInstanceRead(ctx, d, m)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*gomaasapi.MAASObject)

	// Get MAAS machine
	machineManager, err := maas.NewMachineManager(d.Id(), gmaw.NewMachine(client))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get machine (%s) manager: %s", d.Id(), err))
	}
	machine := machineManager.Current()

	// Set Terraform state
	d.Set("cpu_count", machine.CPUCount)
	d.Set("memory", machine.Memory)
	d.Set("fqdn", machine.FQDN)
	ipAddresses := make([]string, len(machine.IPAddresses))
	for i, ip := range machine.IPAddresses {
		ipAddresses[i] = ip.String()
	}
	d.Set("ip_addresses", ipAddresses)

	return nil
}

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*gomaasapi.MAASObject)

	// Release MAAS machine
	machinesManager := maas.NewMachinesManager(gmaw.NewMachines(client))
	err := machinesManager.Release([]string{d.Id()}, "Released by Terraform")
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to release machine (%s): %s", d.Id(), err))
	}

	// Wait MAAS machine to be released
	log.Printf("[DEBUG] Waiting for machine (%s) to be released\n", d.Id())
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Releasing"},
		Target:     []string{"Ready"},
		Refresh:    getMachineStatusFunc(client, d.Id()),
		Timeout:    1 * time.Minute,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("machine (%s) didn't release within allowed timeout: %s", d.Id(), err))
	}

	return nil
}

func getMachinesAllocateParams(d *schema.ResourceData) *endpoint.MachinesAllocateParams {
	allocateParams := endpoint.MachinesAllocateParams{}

	if cpuCount, ok := d.GetOk("min_cpu_count"); ok {
		allocateParams.CPUCount = cpuCount.(int)
	}
	if memory, ok := d.GetOk("min_memory"); ok {
		allocateParams.Mem = memory.(int)
	}
	if tags, ok := d.GetOk("tags"); ok {
		allocateParams.Tags = convertToStringSlice(tags)
	}
	if zone, ok := d.GetOk("zone"); ok {
		allocateParams.Zone = zone.(string)
	}
	if pool, ok := d.GetOk("pool"); ok {
		allocateParams.Pool = pool.(string)
	}

	return &allocateParams
}

func getMachineDeployParams(d *schema.ResourceData) *endpoint.MachineDeployParams {
	deployParams := endpoint.MachineDeployParams{}

	if userData, ok := d.GetOk("user_data"); ok {
		deployParams.UserData = base64Encode([]byte(userData.(string)))
	}
	if distroSeries, ok := d.GetOk("distro_series"); ok {
		deployParams.DistroSeries = distroSeries.(string)
	}
	if kernel, ok := d.GetOk("hwe_kernel"); ok {
		deployParams.HWEKernel = kernel.(string)
	}

	return &deployParams
}

func getMachineStatusFunc(client *gomaasapi.MAASObject, systemId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		machineManager, err := maas.NewMachineManager(systemId, gmaw.NewMachine(client))
		if err != nil {
			log.Printf("[ERROR] Unable to get machine (%s) status: %s\n", systemId, err)
			return nil, "", err
		}
		machine := machineManager.Current()

		log.Printf("[DEBUG] Machine (%s) status: %s\n", systemId, machine.StatusName)
		return machine, machine.StatusName, nil
	}
}
