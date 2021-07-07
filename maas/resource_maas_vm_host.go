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
				vmHost, err := getVMHost(client, d.Id())
				if err != nil {
					return nil, err
				}
				tfState := map[string]interface{}{
					"id":   fmt.Sprintf("%v", vmHost.ID),
					"type": vmHost.Type,
				}
				if vmHost.Host.SystemID != "" {
					tfState["machine"] = vmHost.Host.SystemID
				} else {
					vmHostParams, err := client.VMHost.GetParameters(vmHost.ID)
					if err != nil {
						return nil, err
					}
					for _, k := range []string{"power_address", "power_user", "power_pass"} {
						if val, ok := vmHostParams[k]; ok {
							tfState[k] = val
						}
					}
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"lxd", "virsh"}, false)),
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
				Computed: true,
			},
			"memory_over_commit_ratio": {
				Type:     schema.TypeFloat,
				Optional: true,
				Computed: true,
			},
			"default_macvlan_mode": {
				Type:     schema.TypeString,
				Optional: true,
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
		vmHost, err = client.VMHosts.Create(getVMHostParams(d))
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
	tfState := map[string]interface{}{
		"name":                          vmHost.Name,
		"zone":                          vmHost.Zone.Name,
		"pool":                          vmHost.Pool.Name,
		"tags":                          vmHost.Tags,
		"cpu_over_commit_ratio":         vmHost.CPUOverCommitRatio,
		"memory_over_commit_ratio":      vmHost.MemoryOverCommitRatio,
		"default_macvlan_mode":          vmHost.DefaultMACVLANMode,
		"resources_cores_total":         vmHost.Total.Cores,
		"resources_memory_total":        vmHost.Total.Memory,
		"resources_local_storage_total": vmHost.Total.LocalStorage,
	}
	if err := setTerraformState(d, tfState); err != nil {
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
	_, err = client.VMHost.Update(vmHost.ID, getVMHostParams(d))
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

func getVMHostParams(d *schema.ResourceData) *entity.VMHostParams {
	return &entity.VMHostParams{
		Name:                  d.Get("name").(string),
		Type:                  d.Get("type").(string),
		PowerAddress:          d.Get("power_address").(string),
		PowerUser:             d.Get("power_user").(string),
		PowerPass:             d.Get("power_pass").(string),
		CPUOverCommitRatio:    d.Get("cpu_over_commit_ratio").(float64),
		MemoryOverCommitRatio: d.Get("memory_over_commit_ratio").(float64),
		DefaultMacvlanMode:    d.Get("default_macvlan_mode").(string),
		Zone:                  d.Get("zone").(string),
		Pool:                  d.Get("pool").(string),
		Tags:                  strings.Join(convertToStringSlice(d.Get("tags").(*schema.Set).List()), ","),
	}
}

func deployMachineAsVMHost(ctx context.Context, client *client.Client, machineIdentifier string, vmHostType string) (*entity.VMHost, error) {
	// Find machine
	machine, err := getMachine(client, machineIdentifier)
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

func getVMHost(client *client.Client, identifier string) (*entity.VMHost, error) {
	vmHosts, err := client.VMHosts.Get()
	if err != nil {
		return nil, err
	}
	for _, vmHost := range vmHosts {
		if fmt.Sprintf("%v", vmHost.ID) == identifier || vmHost.Name == identifier {
			return &vmHost, err
		}
	}
	return nil, fmt.Errorf("VM host (%s) not found", identifier)
}
