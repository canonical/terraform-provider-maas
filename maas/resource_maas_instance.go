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
			"allocate_params": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
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
						"hostname": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
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
						"tags": {
							Type:     schema.TypeSet,
							Optional: true,
							ForceNew: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"deploy_params": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"distro_series": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
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
						"install_kvm": {
							Type:     schema.TypeBool,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"fqdn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"pool": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cpu_count": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"ip_addresses": {
				Type:     schema.TypeSet,
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
	if err := d.Set("fqdn", machine.FQDN); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("hostname", machine.Hostname); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("zone", machine.Zone.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pool", machine.Pool.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", machine.TagNames); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("cpu_count", machine.CPUCount); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("memory", machine.Memory); err != nil {
		return diag.FromErr(err)
	}
	ipAddresses := make([]string, len(machine.IPAddresses))
	for i, ip := range machine.IPAddresses {
		ipAddresses[i] = ip.String()
	}
	if err := d.Set("ip_addresses", ipAddresses); err != nil {
		return diag.FromErr(err)
	}

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
	params := endpoint.MachinesAllocateParams{}

	data, ok := d.GetOk("allocate_params")
	if !ok {
		return &params
	}
	ap := data.([]interface{})[0].(map[string]interface{})

	if p, ok := ap["min_cpu_count"]; ok {
		params.CPUCount = p.(int)
	}
	if p, ok := ap["min_memory"]; ok {
		params.Mem = p.(int)
	}
	if p, ok := ap["hostname"]; ok {
		params.Name = p.(string)
	}
	if p, ok := ap["zone"]; ok {
		params.Zone = p.(string)
	}
	if p, ok := ap["pool"]; ok {
		params.Pool = p.(string)
	}
	if p, ok := ap["tags"]; ok {
		params.Tags = convertToStringSlice(p.(*schema.Set).List())
	}

	return &params
}

func getMachineDeployParams(d *schema.ResourceData) *endpoint.MachineDeployParams {
	params := endpoint.MachineDeployParams{}

	data, ok := d.GetOk("deploy_params")
	if !ok {
		return &params
	}
	dp := data.([]interface{})[0].(map[string]interface{})

	if p, ok := dp["distro_series"]; ok {
		params.DistroSeries = p.(string)
	}
	if p, ok := dp["hwe_kernel"]; ok {
		params.HWEKernel = p.(string)
	}
	if p, ok := dp["user_data"]; ok {
		params.UserData = base64Encode([]byte(p.(string)))
	}
	if p, ok := dp["install_kvm"]; ok {
		params.InstallKVM = p.(bool)
	}

	return &params
}
