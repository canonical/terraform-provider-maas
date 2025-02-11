package maas

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMaasVMHostMachine() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS VM host machines.",
		CreateContext: resourceVMHostMachineCreate,
		ReadContext:   resourceVMHostMachineRead,
		UpdateContext: resourceVMHostMachineUpdate,
		DeleteContext: resourceVMHostMachineDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client

				machine, err := getMachine(client, d.Id())
				if err != nil {
					return nil, err
				}
				if machine.VMHost.ID == 0 || machine.VMHost.Name == "" || machine.VMHost.ResourceURI == "" {
					return nil, fmt.Errorf("machine (%s) is not a VM host machine", d.Id())
				}
				tfState := map[string]interface{}{
					"id":      machine.SystemID,
					"vm_host": fmt.Sprintf("%v", machine.VMHost.ID),
					"cores":   machine.CPUCount,
					"memory":  machine.Memory,
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},
		UseJSONNumber: true,

		Schema: map[string]*schema.Schema{
			"cores": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "The number of CPU cores (defaults to 1).",
			},
			"domain": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The VM host machine domain. This is computed if it's not set.",
			},
			"hostname": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The VM host machine hostname. This is computed if it's not set.",
			},
			"memory": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "The VM host machine RAM memory, specified in MB (defaults to 2048).",
			},
			"network_interfaces": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Description: "A list of network interfaces for new the VM host. This argument only works when the VM host is deployed from a registered MAAS machine. Parameters defined below. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"fabric": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The fabric for the network interface.",
						},
						"ip_address": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Static IP configured on the new network interface.",
						},
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The network interface name.",
						},
						"subnet_cidr": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The subnet CIDR for the network interface.",
						},
						"vlan": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The VLAN for the network interface.",
						},
					},
				},
			},
			"pinned_cores": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "List of host CPU cores to pin the VM host machine to. If this is passed, the `cores` parameter is ignored.",
			},
			"pool": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The VM host machine pool. This is computed if it's not set.",
			},
			"storage_disks": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Description: "A list of storage disks for the new VM host. Parameters defined below. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"pool": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The VM host storage pool name.",
						},
						"size_gigabytes": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "The storage disk size, specified in GB.",
						},
					},
				},
			},
			"vm_host": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID or name of the VM host used to compose the new machine.",
			},
			"zone": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The VM host machine zone. This is computed if it's not set.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceVMHostMachineCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Find VM host
	vmHost, err := getVMHost(client, d.Get("vm_host").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// Create VM host machine
	params, err := getVMHostMachineParams(d)
	if err != nil {
		return diag.FromErr(err)
	}
	machine, err := client.VMHost.Compose(vmHost.ID, params)
	if err != nil {
		return diag.FromErr(err)
	}

	// Save system id
	d.SetId(machine.SystemID)

	// Wait for VM host machine to be ready
	_, err = waitForMachineStatus(ctx, client, machine.SystemID, []string{"Commissioning", "Testing"}, []string{"Ready"}, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	// Return updated VM host machine
	return resourceVMHostMachineUpdate(ctx, d, meta)
}

func resourceVMHostMachineRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Get VM host machine
	machine, err := client.Machine.Get(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Set Terraform state
	tfState := map[string]interface{}{
		"hostname": machine.Hostname,
		"domain":   machine.Domain.Name,
		"zone":     machine.Zone.Name,
		"pool":     machine.Pool.Name,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVMHostMachineUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Update VM host machine
	if _, err := client.Machine.Update(d.Id(), getVMHostMachineUpdateParams(d), map[string]interface{}{}); err != nil {
		return diag.FromErr(err)
	}

	return resourceVMHostMachineRead(ctx, d, meta)
}

func resourceVMHostMachineDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Delete VM host machine
	err := client.Machine.Delete(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getVMHostMachineParams(d *schema.ResourceData) (*entity.VMHostMachineParams, error) {
	networkInterfaces, err := getVMHostMachineNetworkInterfaces(d.Get("network_interfaces").([]interface{}))
	if err != nil {
		return nil, err
	}
	params := entity.VMHostMachineParams{
		Hostname:    d.Get("hostname").(string),
		Cores:       d.Get("cores").(int),
		PinnedCores: d.Get("pinned_cores").(int),
		Memory:      int64(d.Get("memory").(int)),
		Interfaces:  networkInterfaces,
		Storage:     getVMHostMachineStorageDisks(d.Get("storage_disks").([]interface{})),
	}
	return &params, nil
}

func getVMHostMachineUpdateParams(d *schema.ResourceData) *entity.MachineParams {
	return &entity.MachineParams{
		Hostname: d.Get("hostname").(string),
		Domain:   d.Get("domain").(string),
		Zone:     d.Get("zone").(string),
		Pool:     d.Get("pool").(string),
	}
}

func getVMHostMachineNetworkInterfaces(networkInterfaces []interface{}) (string, error) {
	vmHostNetworkInterfaces := []string{}
	for _, networkInterface := range networkInterfaces {
		n := networkInterface.(map[string]interface{})
		vlan := n["vlan"].(string)
		subnet := n["subnet_cidr"].(string)
		ip := n["ip_address"].(string)
		if vlan == "" && subnet == "" && ip == "" {
			return "", fmt.Errorf("at least one of the network interface properties (vlan, subnet_cidr, ip_address) is required")
		}
		properties := []string{}
		if fabric := n["fabric"].(string); fabric != "" {
			properties = append(properties, fmt.Sprintf("fabric=%s", fabric))
		}
		if vlan != "" {
			properties = append(properties, fmt.Sprintf("vlan=%s", vlan))
		}
		if subnet != "" {
			properties = append(properties, fmt.Sprintf("subnet_cidr=%s", subnet))
		}
		if ip != "" {
			properties = append(properties, fmt.Sprintf("ip=%s", ip))
		}
		vmHostNetworkInterfaces = append(vmHostNetworkInterfaces, fmt.Sprintf("%s:%s", n["name"].(string), strings.Join(properties, ",")))
	}
	return strings.Join(vmHostNetworkInterfaces, ";"), nil
}

func getVMHostMachineStorageDisks(storageDisks []interface{}) string {
	vmHostStorageDisks := []string{}
	for i, storageDisk := range storageDisks {
		d := storageDisk.(map[string]interface{})
		disk := fmt.Sprintf("disk%d:%d", i, int64(d["size_gigabytes"].(int)))
		if pool := d["pool"].(string); pool != "" {
			disk = fmt.Sprintf("%s(%s)", disk, pool)
		}
		vmHostStorageDisks = append(vmHostStorageDisks, disk)
	}
	return strings.Join(vmHostStorageDisks, ",")
}
