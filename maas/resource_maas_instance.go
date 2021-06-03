package maas

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

func resourceMaasInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInstanceCreate,
		ReadContext:   resourceInstanceRead,
		DeleteContext: resourceInstanceDelete,

		Schema: map[string]*schema.Schema{
			"allocate_min_cpu_count": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"allocate_min_memory": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"allocate_hostname": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"allocate_zone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"allocate_pool": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"allocate_tags": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"deploy_distro_series": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"deploy_hwe_kernel": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"deploy_user_data": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
	client := m.(*client.Client)

	// Allocate MAAS machine
	machine, err := client.Machines.Allocate(getMachinesAllocateParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	// Save system id
	d.SetId(machine.SystemID)

	// Deploy MAAS machine
	machine, err = client.Machine.Deploy(machine.SystemID, getMachineDeployParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait for MAAS machine to be deployed
	_, err = waitForMachineStatus(ctx, client, machine.SystemID, []string{"Deploying"}, []string{"Deployed"})
	if err != nil {
		return diag.FromErr(err)
	}

	// Read MAAS machine info
	return resourceInstanceRead(ctx, d, m)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Get MAAS machine
	machine, err := client.Machine.Get(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

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
	client := m.(*client.Client)

	// Release MAAS machine
	err := client.Machines.Release([]string{d.Id()}, "Released by Terraform")
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait MAAS machine to be released
	_, err = waitForMachineStatus(ctx, client, d.Id(), []string{"Releasing"}, []string{"Ready"})
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getMachinesAllocateParams(d *schema.ResourceData) *entity.MachineAllocateParams {
	params := entity.MachineAllocateParams{}

	if p, ok := d.GetOk("allocate_min_cpu_count"); ok {
		params.CPUCount = p.(int)
	}
	if p, ok := d.GetOk("allocate_min_memory"); ok {
		params.Mem = p.(int)
	}
	if p, ok := d.GetOk("allocate_hostname"); ok {
		params.Name = p.(string)
	}
	if p, ok := d.GetOk("allocate_zone"); ok {
		params.Zone = p.(string)
	}
	if p, ok := d.GetOk("allocate_pool"); ok {
		params.Pool = p.(string)
	}
	if p, ok := d.GetOk("allocate_tags"); ok {
		params.Tags = convertToStringSlice(p.(*schema.Set).List())
	}

	return &params
}

func getMachineDeployParams(d *schema.ResourceData) *entity.MachineDeployParams {
	params := entity.MachineDeployParams{}

	if p, ok := d.GetOk("deploy_distro_series"); ok {
		params.DistroSeries = p.(string)
	}
	if p, ok := d.GetOk("deploy_hwe_kernel"); ok {
		params.HWEKernel = p.(string)
	}
	if p, ok := d.GetOk("deploy_user_data"); ok {
		params.UserData = base64Encode([]byte(p.(string)))
	}

	return &params
}
