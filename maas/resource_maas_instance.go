package maas

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

func resourceMaasInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceInstanceCreate,
		ReadContext:   resourceInstanceRead,
		DeleteContext: resourceInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				client := m.(*client.Client)
				machine, err := getMachine(client, d.Id())
				if err != nil {
					return nil, err
				}
				if machine.StatusName != "Deployed" {
					return nil, fmt.Errorf("machine '%s' needs to be already deployed to be imported as maas_instance resource", machine.Hostname)
				}
				d.SetId(machine.SystemID)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"allocate_params": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Description: "Nested argument with the constraints used to machine allocation. Defined below.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_cpu_count": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "The minimum number of cores used to allocate the MAAS machine.",
						},
						"min_memory": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							Description: "The minimum RAM memory size (in MB) used to allocate the MAAS machine.",
						},
						"hostname": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The hostname of the MAAS machine to be allocated.",
						},
						"zone": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The zone name of the MAAS machine to be allocated.",
						},
						"pool": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The pool name of the MAAS machine to be allocated.",
						},
						"tags": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "A set of tag names that must be assigned on the MAAS machine to be allocated.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"deploy_params": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Description: "Nested argument with the config used to deploy the allocated machine. Defined below.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"distro_series": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The distro series used to deploy the allocated MAAS machine. If it's not given, the MAAS server default value is used.",
						},
						"hwe_kernel": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Hardware enablement kernel to use with the image. Only used when deploying Ubuntu.",
						},
						"user_data": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Cloud-init user data script that gets run on the machine once it has deployed. A good practice is to set this with `file(\"/tmp/user-data.txt\")`, where `/tmp/user-data.txt` is a cloud-init script.",
						},
					},
				},
			},
			"network_interfaces": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Description: "Specifies a network interface configuration done before the machine is deployed. Parameters defined below. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the network interface to be configured on the allocated machine.",
						},
						"subnet_cidr": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "An existing subnet CIDR used to configure the network interface. Unless `ip_address` is defined, a free IP address is allocated from the subnet.",
						},
						"ip_address": {
							Type:             schema.TypeString,
							Optional:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
							Description:      "Static IP address to be configured on the network interface. If this is set, the `subnet_cidr` is required.",
						},
					},
				},
			},
			"fqdn": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The deployed MAAS machine FQDN.",
			},
			"hostname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The deployed MAAS machine hostname.",
			},
			"zone": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The deployed MAAS machine zone name.",
			},
			"pool": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The deployed MAAS machine pool name.",
			},
			"tags": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of tag names associated to the deployed MAAS machine.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cpu_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of CPU cores of the deployed MAAS machine.",
			},
			"memory": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The RAM memory size (in GiB) of the deployed MAAS machine.",
			},
			"ip_addresses": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of IP addressed assigned to the deployed MAAS machine.",
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

	// Configure network interfaces
	err = configureInstanceNetworkInterfaces(client, d, machine)
	if err != nil {
		return diag.FromErr(err)
	}

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
	ipAddresses := make([]string, len(machine.IPAddresses))
	for i, ip := range machine.IPAddresses {
		ipAddresses[i] = ip.String()
	}
	tfState := map[string]interface{}{
		"fqdn":         machine.FQDN,
		"hostname":     machine.Hostname,
		"zone":         machine.Zone.Name,
		"pool":         machine.Pool.Name,
		"tags":         machine.TagNames,
		"cpu_count":    machine.CPUCount,
		"memory":       machine.Memory,
		"ip_addresses": ipAddresses,
	}
	if err := setTerraformState(d, tfState); err != nil {
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
	p, ok := d.GetOk("allocate_params")
	if !ok {
		return &entity.MachineAllocateParams{}
	}
	allocateParams := p.(*schema.Set).List()[0].(map[string]interface{})
	return &entity.MachineAllocateParams{
		CPUCount: allocateParams["min_cpu_count"].(int),
		Mem:      allocateParams["min_memory"].(int),
		Name:     allocateParams["hostname"].(string),
		Zone:     allocateParams["zone"].(string),
		Pool:     allocateParams["pool"].(string),
		Tags:     convertToStringSlice(allocateParams["tags"].(*schema.Set).List()),
	}
}

func getMachineDeployParams(d *schema.ResourceData) *entity.MachineDeployParams {
	p, ok := d.GetOk("deploy_params")
	if !ok {
		return &entity.MachineDeployParams{}
	}
	deployParams := p.(*schema.Set).List()[0].(map[string]interface{})
	return &entity.MachineDeployParams{
		DistroSeries: deployParams["distro_series"].(string),
		HWEKernel:    deployParams["hwe_kernel"].(string),
		UserData:     base64Encode([]byte(deployParams["user_data"].(string))),
	}
}

func configureInstanceNetworkInterfaces(client *client.Client, d *schema.ResourceData, machine *entity.Machine) error {
	for _, networkInterface := range d.Get("network_interfaces").(*schema.Set).List() {
		n := networkInterface.(map[string]interface{})
		// Find the machine network interface
		name := n["name"].(string)
		nic, err := getNetworkInterface(client, machine.SystemID, name)
		if err != nil {
			return err
		}
		// Validate the given network configs
		subnetCIDR := n["subnet_cidr"].(string)
		ipAddress := n["ip_address"].(string)
		if subnetCIDR == "" {
			if ipAddress != "" {
				return fmt.Errorf("network interface (%s): 'subnet_cidr' is required when 'ip_address' is set", name)
			}
			// Clear existing network interface links
			// This will leave the network interface disconnected
			if _, err := client.NetworkInterface.Disconnect(machine.SystemID, nic.ID); err != nil {
				return err
			}
			continue
		}
		// Find the subnet
		subnet, err := getSubnet(client, subnetCIDR)
		if err != nil {
			return err
		}
		// Clear existing network interface links
		if _, err := client.NetworkInterface.Disconnect(machine.SystemID, nic.ID); err != nil {
			return err
		}
		// Create new network interface link
		mode := "AUTO"
		if ipAddress != "" {
			mode = "STATIC"
		}
		params := entity.NetworkInterfaceLinkParams{
			Mode:      mode,
			Subnet:    subnet.ID,
			IPAddress: ipAddress,
		}
		if _, err = client.NetworkInterface.LinkSubnet(machine.SystemID, nic.ID, &params); err != nil {
			return err
		}
	}
	return nil
}
