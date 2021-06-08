package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

var (
	vmHostSources = []string{
		"machine",
		"power_address",
	}
	defaultCPUOverCommitRatio    = 1.0
	defaultMemoryOverCommitRatio = 1.0
)

func resourceMaasVMHost() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVMHostCreate,
		ReadContext:   resourceVMHostRead,
		UpdateContext: resourceVMHostUpdate,
		DeleteContext: resourceVMHostDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				client := m.(*client.Client)
				vmHost, err := findVMHost(client, d.Id())
				if err != nil {
					return nil, err
				}
				d.SetId(fmt.Sprintf("%v", vmHost.ID))
				if err := d.Set("type", vmHost.Type); err != nil {
					return nil, err
				}
				if err := d.Set("cpu_over_commit_ratio", defaultCPUOverCommitRatio); err != nil {
					return nil, err
				}
				if err := d.Set("memory_over_commit_ratio", defaultMemoryOverCommitRatio); err != nil {
					return nil, err
				}
				if vmHost.Host.SystemID != "" {
					if err := d.Set("machine", vmHost.Host.SystemID); err != nil {
						return nil, err
					}
				} else {
					vmHostParams, err := client.VMHost.GetParameters(vmHost.ID)
					if err != nil {
						return nil, err
					}
					if err := d.Set("power_address", vmHostParams.PowerAddress); err != nil {
						return nil, err
					}
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateDiagFunc: validation.ToDiagFunc(
					validation.StringInSlice([]string{"lxd", "virsh"}, false)),
			},
			"machine": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ExactlyOneOf:  vmHostSources,
				ConflictsWith: []string{"power_address", "power_user", "power_pass"},
			},
			"power_address": {
				Type:          schema.TypeString,
				Optional:      true,
				ExactlyOneOf:  vmHostSources,
				ConflictsWith: []string{"machine"},
			},
			"power_user": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"machine"},
			},
			"power_pass": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"machine"},
			},
			"name": {
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
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cpu_over_commit_ratio": {
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  defaultCPUOverCommitRatio,
			},
			"memory_over_commit_ratio": {
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  defaultMemoryOverCommitRatio,
			},
			"default_macvlan_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"resources_cores_available": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"resources_memory_available": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"resources_local_storage_available": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"resources_cores_total": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"resources_memory_total": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"resources_local_storage_total": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourceVMHostCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Create VM host
	var vmHost *entity.VMHost
	var err error
	if p, ok := d.GetOk("machine"); ok {
		// Deploy machine, and register it as VM host
		vmHost, err = deployMachineAsVMHost(ctx, client, p.(string), d.Get("type").(string))
		if err != nil {
			return diag.FromErr(err)
		}
	} else {
		vmHost, err = client.VMHosts.Create(getVMHostCreateParams(d))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	// Save Id
	d.SetId(fmt.Sprintf("%v", vmHost.ID))

	// Return updated VM host
	return resourceVMHostUpdate(ctx, d, m)
}

func resourceVMHostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Get VM host details
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	vmHost, err := client.VMHost.Get(id)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set Terraform state
	if err := d.Set("name", vmHost.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("zone", vmHost.Zone.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pool", vmHost.Pool.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", vmHost.Tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_cores_available", vmHost.Available.Cores); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_cores_total", vmHost.Total.Cores); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_memory_available", vmHost.Available.Memory); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_memory_total", vmHost.Total.Memory); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_local_storage_available", vmHost.Available.LocalStorage); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_local_storage_total", vmHost.Total.LocalStorage); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVMHostUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Get the VM host
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	vmHost, err := client.VMHost.Get(id)
	if err != nil {
		return diag.FromErr(err)
	}

	// Update VM host options
	vmHostParams, err := client.VMHost.GetParameters(vmHost.ID)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.VMHost.Update(vmHost.ID, getVMHostUpdateParams(d, vmHost, vmHostParams))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceVMHostRead(ctx, d, m)
}

func resourceVMHostDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Delete VM host
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	vmHost, err := client.VMHost.Get(id)
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.VMHost.Delete(vmHost.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	// If the VM host was deployed from a machine, release the machine.
	if vmHost.Host.SystemID != "" {
		// Release machine
		err = client.Machines.Release([]string{vmHost.Host.SystemID}, "Released by Terraform")
		if err != nil {
			return diag.FromErr(err)
		}
		// Wait machine to be released
		_, err = waitForMachineStatus(ctx, client, vmHost.Host.SystemID, []string{"Releasing"}, []string{"Ready"})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func getVMHostCreateParams(d *schema.ResourceData) *entity.VMHostParams {
	params := entity.VMHostParams{
		Type:                  d.Get("type").(string),
		CPUOverCommitRatio:    d.Get("cpu_over_commit_ratio").(float64),
		MemoryOverCommitRatio: d.Get("memory_over_commit_ratio").(float64),
	}

	if p, ok := d.GetOk("power_address"); ok {
		params.PowerAddress = p.(string)
	}
	if p, ok := d.GetOk("power_user"); ok {
		params.PowerUser = p.(string)
	}
	if p, ok := d.GetOk("power_pass"); ok {
		params.PowerPass = p.(string)
	}

	return &params
}

func getVMHostUpdateParams(d *schema.ResourceData, vmHost *entity.VMHost, params *entity.VMHostParams) *entity.VMHostParams {
	params.Type = vmHost.Type
	params.Name = vmHost.Name
	params.CPUOverCommitRatio = vmHost.CPUOverCommitRatio
	params.MemoryOverCommitRatio = vmHost.MemoryOverCommitRatio
	params.DefaultMacvlanMode = vmHost.DefaultMACVLANMode
	params.Zone = vmHost.Zone.Name
	params.Pool = vmHost.Pool.Name
	params.Tags = strings.Join(vmHost.Tags, ",")

	if p, ok := d.GetOk("power_address"); ok {
		params.PowerAddress = p.(string)
	}
	if p, ok := d.GetOk("power_pass"); ok {
		params.PowerPass = p.(string)
	}
	if p, ok := d.GetOk("name"); ok {
		params.Name = p.(string)
	}
	if p, ok := d.GetOk("zone"); ok {
		params.Zone = p.(string)
	}
	if p, ok := d.GetOk("pool"); ok {
		params.Pool = p.(string)
	}
	if p, ok := d.GetOk("tags"); ok {
		params.Tags = strings.Join(convertToStringSlice(p.(*schema.Set).List()), ",")
	}
	if p, ok := d.GetOk("cpu_over_commit_ratio"); ok {
		params.CPUOverCommitRatio = p.(float64)
	}
	if p, ok := d.GetOk("memory_over_commit_ratio"); ok {
		params.MemoryOverCommitRatio = p.(float64)
	}
	if p, ok := d.GetOk("default_macvlan_mode"); ok {
		params.DefaultMacvlanMode = p.(string)
	}

	return params
}

func deployMachineAsVMHost(ctx context.Context, client *client.Client, machineIdentifier string, vmHostType string) (*entity.VMHost, error) {
	// Find machine
	machine, err := findMachine(client, machineIdentifier)
	if err != nil {
		return nil, err
	}

	// Allocate machine
	allocateParams := entity.MachineAllocateParams{SystemID: machine.SystemID}
	machine, err = client.Machines.Allocate(&allocateParams)
	if err != nil {
		return nil, err
	}

	// Deploy machine
	deployParams := entity.MachineDeployParams{
		DistroSeries:   "focal",
		InstallKVM:     (vmHostType == "virsh"),
		RegisterVMHost: (vmHostType == "lxd"),
	}
	machine, err = client.Machine.Deploy(machine.SystemID, &deployParams)
	if err != nil {
		return nil, err
	}

	// Wait for MAAS machine to be deployed
	machine, err = waitForMachineStatus(ctx, client, machine.SystemID, []string{"Deploying"}, []string{"Deployed"})
	if err != nil {
		return nil, err
	}

	// Return the VM host
	vmHosts, err := client.VMHosts.Get()
	if err != nil {
		return nil, err
	}
	for _, vmHost := range vmHosts {
		if vmHost.Host.SystemID == machine.SystemID {
			return &vmHost, nil
		}
	}

	return nil, fmt.Errorf("cannot find registered VM host on machine '%s'", machineIdentifier)
}
