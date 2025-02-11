package maas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/juju/gomaasapi/v2"
)

var (
	vmHostSources = []string{
		"machine",
		"power_address",
	}
)

func resourceMaasVMHost() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS VM hosts.",
		CreateContext: resourceVMHostCreate,
		ReadContext:   resourceVMHostRead,
		UpdateContext: resourceVMHostUpdate,
		DeleteContext: resourceVMHostDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client

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
			"cpu_over_commit_ratio": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Computed:    true,
				Description: "The new VM host CPU overcommit ratio. This is computed if it's not set.",
			},
			"default_macvlan_mode": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The new VM host default macvlan mode. Supported values are: `bridge`, `passthru`, `private`, `vepa`. This is computed if it's not set.",
			},
			"machine": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ExactlyOneOf:  vmHostSources,
				ConflictsWith: []string{"power_address", "power_user", "power_pass"},
				Description:   "The identifier (hostname, FQDN or system ID) of a registered ready MAAS machine. This is going to be deployed and registered as a new VM host. This argument conflicts with: `power_address`, `power_user`, `power_pass`.",
			},
			"memory_over_commit_ratio": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Computed:    true,
				Description: "The new VM host RAM memory overcommit ratio. This is computed if it's not set.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The new VM host name. This is computed if it's not set.",
			},
			"pool": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The new VM host pool name. This is computed if it's not set.",
			},
			"power_address": {
				Type:          schema.TypeString,
				Optional:      true,
				ExactlyOneOf:  vmHostSources,
				ConflictsWith: []string{"machine"},
				Description:   "Address that gives MAAS access to the VM host power control. For example: `qemu+ssh://172.16.99.2/system`. The address given here must reachable by the MAAS server. It can't be set if `machine` argument is used.",
			},
			"power_pass": {
				Type:          schema.TypeString,
				Optional:      true,
				Sensitive:     true,
				ConflictsWith: []string{"machine"},
				Description:   "User password to use for power control of the VM host. Cannot be set if `machine` parameter is used.",
			},
			"power_user": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"machine"},
				Description:   "User name to use for power control of the VM host. Cannot be set if `machine` parameter is used.",
			},
			"resources_cores_total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The VM host total number of CPU cores.",
			},
			"resources_local_storage_total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The VM host total local storage (in bytes).",
			},
			"resources_memory_total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The VM host total RAM memory (in MB).",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Computed:    true,
				Description: "A set of tag names to assign to the new VM host. This is computed if it's not set.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"lxd", "virsh"}, false)),
				Description:      "The VM host type. Supported values are: `lxd`, `virsh`.",
			},
			"zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The new VM host zone name. This is computed if it's not set.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceVMHostCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Create VM host
	var vmHost *entity.VMHost
	var err error
	if p, ok := d.GetOk("machine"); ok {
		// Deploy machine, and register it as VM host
		vmHost, err = deployMachineAsVMHost(ctx, client, p.(string), d.Get("type").(string), d.Timeout(schema.TimeoutCreate))
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
	return resourceVMHostUpdate(ctx, d, meta)
}

func resourceVMHostRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

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

func resourceVMHostUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Get the VM host
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Update VM host options
	_, err = client.VMHost.Update(id, getVMHostParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceVMHostRead(ctx, d, meta)
}

func resourceVMHostDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

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

	// Check if VM host was linked to a dynamic machine and if yes, return
	// Dynamic machines are deleted by MAAS when their VM hosts are deleted.
	// This information is not directly available from the API.
	_, err = client.Machine.Get(vmHost.Host.SystemID)
	if err, ok := gomaasapi.GetServerError(err); ok {
		if err.StatusCode == http.StatusNotFound {
			return nil
		}
	}

	// VM host was deployed from a machine, so release the machine.
	err = client.Machines.Release([]string{vmHost.Host.SystemID}, "Released by Terraform")
	if err != nil {
		return diag.FromErr(err)
	}
	// Wait machine to be released
	_, err = waitForMachineStatus(ctx, client, vmHost.Host.SystemID, []string{"Releasing"}, []string{"Ready"}, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
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

func deployMachineAsVMHost(ctx context.Context, client *client.Client, machineIdentifier string, vmHostType string, maxTimeout time.Duration) (*entity.VMHost, error) {
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

	// Get Default OS and series
	var defaultOsystem string
	defaultOsystemBytes, err := client.MAASServer.Get("default_osystem")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(defaultOsystemBytes, &defaultOsystem)
	if err != nil {
		return nil, err
	}
	var defaultSeries string
	defaultSeriesBytes, err := client.MAASServer.Get("default_distro_series")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(defaultSeriesBytes, &defaultSeries)
	if err != nil {
		return nil, err
	}

	// Deploy machine
	deployParams := entity.MachineDeployParams{
		DistroSeries:   fmt.Sprintf("%s/%s", defaultOsystem, defaultSeries),
		InstallKVM:     (vmHostType == "virsh"),
		RegisterVMHost: (vmHostType == "lxd"),
	}
	machine, err = client.Machine.Deploy(machine.SystemID, &deployParams)
	if err != nil {
		return nil, err
	}

	// Wait for MAAS machine to be deployed
	machine, err = waitForMachineStatus(ctx, client, machine.SystemID, []string{"Deploying"}, []string{"Deployed"}, maxTimeout)
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
