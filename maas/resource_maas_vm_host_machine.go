package maas

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

func resourceMaasVMHostMachine() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVMHostMachineCreate,
		ReadContext:   resourceVMHostMachineRead,
		UpdateContext: resourceVMHostMachineUpdate,
		DeleteContext: resourceVMHostMachineDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				client := m.(*client.Client)
				machine, err := findMachine(client, d.Id())
				if err != nil {
					return nil, err
				}
				if machine.VMHost.ID == 0 || machine.VMHost.Name == "" || machine.VMHost.ResourceURI == "" {
					return nil, fmt.Errorf("machine (%s) is not a VM host machine", d.Id())
				}
				d.SetId(machine.SystemID)
				if err := d.Set("vm_host", fmt.Sprintf("%v", machine.VMHost.ID)); err != nil {
					return nil, err
				}
				if err := d.Set("cores", machine.CPUCount); err != nil {
					return nil, err
				}
				if err := d.Set("memory", machine.Memory); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"vm_host": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cores": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  1,
			},
			"pinned_cores": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Optional: true,
				ForceNew: true,
				Default:  2048,
			},
			"network_interfaces": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"fabric": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"vlan": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"subnet_cidr": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ip_address": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"storage_disks": {
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size_gigabytes": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"pool": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"domain": {
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
		},
	}
}

func resourceVMHostMachineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Find VM host
	vmHost, err := findVMHost(client, d.Get("vm_host").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// Create VM host machine
	params, err := getVMHostMachineCreateParams(d)
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
	_, err = waitForMachineStatus(ctx, client, machine.SystemID, []string{"Commissioning", "Testing"}, []string{"Ready"})
	if err != nil {
		return diag.FromErr(err)
	}

	// Return updated VM host machine
	return resourceVMHostMachineUpdate(ctx, d, m)
}

func resourceVMHostMachineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Get VM host machine
	machine, err := client.Machine.Get(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Set Terraform state
	if err := d.Set("hostname", machine.Hostname); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("domain", machine.Domain.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("zone", machine.Zone.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pool", machine.Pool.Name); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVMHostMachineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Update VM host machine
	machine, err := client.Machine.Get(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = client.Machine.Update(machine.SystemID, getVMHostMachineUpdateParams(d, machine), map[string]string{})
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceVMHostMachineRead(ctx, d, m)
}

func resourceVMHostMachineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Delete VM host machine
	err := client.Machine.Delete(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func findVMHost(client *client.Client, vmHostIdentifier string) (*entity.VMHost, error) {
	vmHosts, err := client.VMHosts.Get()
	if err != nil {
		return nil, err
	}

	for _, vmHost := range vmHosts {
		if fmt.Sprintf("%v", vmHost.ID) == vmHostIdentifier || vmHost.Name == vmHostIdentifier {
			return &vmHost, err
		}
	}

	return nil, fmt.Errorf("VM host (%s) not found", vmHostIdentifier)
}

func getVMHostMachineCreateParams(d *schema.ResourceData) (*entity.VMHostMachineParams, error) {
	params := entity.VMHostMachineParams{}

	if p, ok := d.GetOk("cores"); ok {
		params.Cores = p.(int)
	}
	if p, ok := d.GetOk("pinned_cores"); ok {
		params.PinnedCores = p.(int)
	}
	if p, ok := d.GetOk("memory"); ok {
		params.Memory = p.(int)
	}
	if p, ok := d.GetOk("network_interfaces"); ok {
		networkInterfaces, err := getVMHostMachineNetworkInterfaces(p.([]interface{}))
		if err != nil {
			return nil, err
		}
		params.Interfaces = networkInterfaces
	}
	if p, ok := d.GetOk("storage_disks"); ok {
		params.Storage = getVMHostMachineStorageDisks(p.([]interface{}))
	}
	if p, ok := d.GetOk("hostname"); ok {
		params.Hostname = p.(string)
	}

	return &params, nil
}

func getVMHostMachineUpdateParams(d *schema.ResourceData, machine *entity.Machine) *entity.MachineParams {
	params := entity.MachineParams{
		CPUCount:     machine.CPUCount,
		Memory:       machine.Memory,
		SwapSize:     machine.SwapSize,
		Architecture: machine.Architecture,
		MinHWEKernel: machine.MinHWEKernel,
		PowerType:    machine.PowerType,
		Description:  machine.Description,
	}

	if p, ok := d.GetOk("hostname"); ok {
		params.Hostname = p.(string)
	}
	if p, ok := d.GetOk("domain"); ok {
		params.Domain = p.(string)
	}
	if p, ok := d.GetOk("zone"); ok {
		params.Zone = p.(string)
	}
	if p, ok := d.GetOk("pool"); ok {
		params.Pool = p.(string)
	}

	return &params
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
		disk := fmt.Sprintf("disk%d:%d", i, d["size_gigabytes"].(int))
		if pool := d["pool"].(string); pool != "" {
			disk = fmt.Sprintf("%s(%s)", disk, pool)
		}
		vmHostStorageDisks = append(vmHostStorageDisks, disk)
	}
	return strings.Join(vmHostStorageDisks, ",")
}
