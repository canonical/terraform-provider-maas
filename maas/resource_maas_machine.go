package maas

import (
	"context"
	"fmt"
	"log"
	"math"
	"reflect"
	"sort"
	"time"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceMAASMachine() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS machines.",
		CreateContext: resourceMachineCreate,
		ReadContext:   resourceMachineRead,
		UpdateContext: resourceMachineUpdate,
		DeleteContext: resourceMachineDelete,
		SchemaVersion: 1,
		StateUpgraders: []schema.StateUpgrader{
			{
				Type:    resourceMAASMachineResourceV0().CoreConfigSchema().ImpliedType(),
				Upgrade: resourceMAASMachineStateUpgradeV0,
				Version: 0,
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client

				machine, err := getMachine(client, d.Id())
				if err != nil {
					return nil, err
				}

				powerParams, err := client.Machine.GetPowerParameters(machine.SystemID)
				if err != nil {
					return nil, err
				}

				powerParamsString, err := structure.FlattenJsonToString(powerParams)
				if err != nil {
					return nil, err
				}

				tfState := map[string]any{
					"id":               machine.SystemID,
					"power_type":       machine.PowerType,
					"power_parameters": powerParamsString,
					"pxe_mac_address":  machine.BootInterface.MACAddress,
					"architecture":     machine.Architecture,
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}

				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"architecture": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "amd64/generic",
				Description: "The architecture type of the machine. Defaults to `amd64/generic`.",
			},
			"block_devices": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "A list of block devices attached to the machine.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id_path": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID path of the block device.",
						},
						"model": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The model of the block device.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The block device name.",
						},
						"size_gigabytes": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The size of the block device (in GB).",
						},
					},
				},
			},
			"commissioning_scripts": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Commissioning script names and tags to be run. By default all custom commissioning scripts are run. Built-in commissioning scripts always run. Selecting 'update_firmware' or 'configure_hba' will run firmware updates or configure HBA's on matching machines.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The domain of the machine. This is computed if it's not set.",
			},
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The machine hostname. This is computed if it's not set.",
			},
			"min_hwe_kernel": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The minimum kernel version allowed to run on this machine. Only used when deploying Ubuntu. This is computed if it's not set.",
			},
			"network_interfaces": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of MAC addresses of network interfaces attached to the machine.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"pool": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The resource pool of the machine. This is computed if it's not set.",
			},
			"power_parameters": {
				Type:         schema.TypeString,
				Required:     true,
				Sensitive:    true,
				ValidateFunc: validation.StringIsJSON,
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					oldMap, err := structure.ExpandJsonFromString(oldValue)
					if err != nil {
						return false
					}

					newMap, err := structure.ExpandJsonFromString(newValue)
					if err != nil {
						return false
					}

					return reflect.DeepEqual(oldMap, newMap)
				},
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				Description: "Serialized JSON string containing the parameters specific to the `power_type`. See [Power types](https://maas.io/docs/api#power-types) section for a list of the available power parameters for each power type.",
			},
			"power_type": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A power management type (e.g. `ipmi`).",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(
					[]string{
						"amt", "apc", "dli", "eaton", "hmc", "ipmi", "manual", "moonshot",
						"mscm", "msftocs", "nova", "openbmc", "proxmox", "recs_box", "redfish",
						"sm15k", "ucsm", "vmware", "webhook", "wedge", "lxd", "virsh",
					},
					false)),
			},
			"pxe_mac_address": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The MAC address of the machine's PXE boot NIC.",
			},
			"script_parameters": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Scripts specified to run may define their own parameters. These parameters may be passed as parameter name (key) value pairs as a map. Optionally a parameter may have the script name prepended to have that parameter only apply to that specific script, e.g. my-script_param=value.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"testing_scripts": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Testing scripts names and tags to be run after commissioning. By default all tests tagged 'testing' will be run. Set to ['none'] to disable running tests.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The zone of the machine. This is computed if it's not set.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceMachineCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Create MAAS machine
	powerParams, err := getMachinePowerParams(d)
	if err != nil {
		return diag.FromErr(err)
	}

	machine, err := client.Machines.Create(getMachineCreateParams(d), powerParams)
	if err != nil {
		// Clean up trailing resources, as the gomaasclient does not return the created machine on error
		badMachine, errDel := getMachine(client, d.Get("pxe_mac_address").(string))
		if errDel != nil {
			return diag.FromErr(fmt.Errorf("error creating MAAS machine: %v;\nAdditionally, error when attempting to get the trailing resource: %v", err, errDel))
		}

		errDel = client.Machine.Delete(badMachine.SystemID)
		if errDel != nil {
			return diag.FromErr(fmt.Errorf("error creating MAAS machine: %v;\nAdditionally, error when attempting to delete the trailing resource: %v", err, errDel))
		}

		return diag.FromErr(err)
	}

	// Save Id
	d.SetId(machine.SystemID)

	// Wait for machine to be ready
	_, err = waitForMachineStatus(ctx, client, machine.SystemID, []string{"Commissioning", "Testing"}, []string{"Ready"}, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	// Read machine info
	return resourceMachineRead(ctx, d, meta)
}

func resourceMachineRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Get machine
	machine, err := client.Machine.Get(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Set Terraform state
	tfState := map[string]any{
		"architecture":   machine.Architecture,
		"min_hwe_kernel": machine.MinHWEKernel,
		"hostname":       machine.Hostname,
		"domain":         machine.Domain.Name,
		"zone":           machine.Zone.Name,
		"pool":           machine.Pool.Name,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	networkInterfaces := make([]string, len(machine.InterfaceSet))
	for i, networkInterface := range machine.InterfaceSet {
		networkInterfaces[i] = networkInterface.MACAddress
	}

	if err := d.Set("network_interfaces", networkInterfaces); err != nil {
		return diag.FromErr(err)
	}

	blockDevices := getAllBlockDeviceMachineParameters(machine.BlockDeviceSet)
	if err := d.Set("block_devices", blockDevices); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceMachineUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	scriptsHaveChanged := d.HasChanges("commissioning_scripts", "testing_scripts", "script_parameters")
	// Update machine
	machine, err := client.Machine.Get(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	powerParams, err := getMachinePowerParams(d)
	if err != nil {
		return diag.FromErr(err)
	}

	_, err = client.Machine.Update(machine.SystemID, getMachineUpdateParams(d), powerParams)
	if err != nil {
		return diag.FromErr(err)
	}

	if scriptsHaveChanged {
		machine, err = client.Machine.Commission(machine.SystemID, getMachineCommissionParams(d))
		if err != nil {
			return diag.FromErr(err)
		}

		// Wait for machine to be ready
		_, err = waitForMachineStatus(ctx, client, machine.SystemID, []string{"Commissioning", "Testing"}, []string{"Ready"}, d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceMachineRead(ctx, d, meta)
}

func resourceMachineDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Delete machine
	if err := client.Machine.Delete(d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getMachinePowerParams(d *schema.ResourceData) (map[string]any, error) {
	powerParams := make(map[string]any)
	powerParamsString := d.Get("power_parameters").(string)

	params, err := structure.ExpandJsonFromString(powerParamsString)
	if err != nil {
		return powerParams, err
	}

	for k, v := range params {
		powerParams[fmt.Sprintf("power_parameters_%s", k)] = v
	}

	return powerParams, nil
}

func getMachineCreateParams(d *schema.ResourceData) *entity.MachineCreateParams {
	return &entity.MachineCreateParams{
		Commission:           true,
		PowerType:            d.Get("power_type").(string),
		MACAddresses:         []string{d.Get("pxe_mac_address").(string)},
		Architecture:         d.Get("architecture").(string),
		MinHWEKernel:         d.Get("min_hwe_kernel").(string),
		Hostname:             d.Get("hostname").(string),
		Domain:               d.Get("domain").(string),
		Zone:                 d.Get("zone").(string),
		Pool:                 d.Get("pool").(string),
		CommissioningScripts: listAsStringBase(d.Get("commissioning_scripts").([]any)),
		TestingScripts:       listAsStringBase(d.Get("testing_scripts").([]any)),
		ScriptParams:         d.Get("script_parameters").(map[string]any),
	}
}

func getMachineUpdateParams(d *schema.ResourceData) *entity.MachineUpdateParams {
	return &entity.MachineUpdateParams{
		Commission:   true,
		PowerType:    d.Get("power_type").(string),
		MACAddresses: []string{d.Get("pxe_mac_address").(string)},
		Architecture: d.Get("architecture").(string),
		MinHWEKernel: d.Get("min_hwe_kernel").(string),
		Hostname:     d.Get("hostname").(string),
		Domain:       d.Get("domain").(string),
		Zone:         d.Get("zone").(string),
		Pool:         d.Get("pool").(string),
	}
}

func getMachineCommissionParams(d *schema.ResourceData) *entity.MachineCommissionParams {
	return &entity.MachineCommissionParams{
		CommissioningScripts: listAsStringBase(d.Get("commissioning_scripts").([]any)),
		TestingScripts:       listAsStringBase(d.Get("testing_scripts").([]any)),
		ScriptParams:         d.Get("script_parameters").(map[string]any),
	}
}

func getMachineStatusFunc(client *client.Client, systemID string) retry.StateRefreshFunc {
	return func() (any, string, error) {
		machine, err := client.Machine.Get(systemID)
		if err != nil {
			return nil, "", err
		}

		log.Printf("[DEBUG] Machine (%s) status: %s\n", systemID, machine.StatusName)

		return machine, machine.StatusName, nil
	}
}

func waitForMachineStatus(ctx context.Context, client *client.Client, systemID string, pendingStates []string, targetStates []string, maxTimeout time.Duration) (*entity.Machine, error) {
	log.Printf("[DEBUG] Waiting for machine (%s) status to be one of %s\n", systemID, targetStates)
	stateConf := &retry.StateChangeConf{
		Pending:    pendingStates,
		Target:     targetStates,
		Refresh:    getMachineStatusFunc(client, systemID),
		Timeout:    maxTimeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	result, err := stateConf.WaitForStateContext(ctx)
	if err != nil {
		return nil, err
	}

	return result.(*entity.Machine), nil
}

func getMachine(client *client.Client, identifier string) (*entity.Machine, error) {
	machines, err := client.Machines.Get(&entity.MachinesParams{})
	if err != nil {
		return nil, err
	}

	for _, m := range machines {
		if m.SystemID == identifier || m.Hostname == identifier || m.FQDN == identifier || m.BootInterface.MACAddress == identifier {
			return &m, nil
		}
	}

	return nil, fmt.Errorf("machine (%s) not found", identifier)
}

func getAllBlockDeviceMachineParameters(blockDevices []entity.BlockDevice) []map[string]any {
	// sort block devices by ID
	sort.Slice(blockDevices, func(i, j int) bool {
		return blockDevices[i].ID < blockDevices[j].ID
	})

	// Create a slice of maps to hold block device parameters
	blockDeviceParams := make([]map[string]any, len(blockDevices))
	for i, blockDevice := range blockDevices {
		blockDeviceParams[i] = map[string]any{
			"name":           blockDevice.Name,
			"size_gigabytes": int(math.Round(float64(blockDevice.Size) / GigaBytes)),
			"id_path":        blockDevice.IDPath,
			"model":          blockDevice.Model,
		}
	}

	return blockDeviceParams
}
