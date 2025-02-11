package maas

import (
	"context"
	"fmt"
	"time"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceMaasInstance() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to deploy and release machines already configured in MAAS, based on the specified parameters. If no parameters are given, a random machine will be allocated and deployed using the defaults.\n\n**NOTE:** The MAAS provider currently provides both standalone resources and in-line resources for network interfaces. You cannot use in-line network interfaces in conjunction with any standalone network interfaces resources. Doing so will cause conflicts and will overwrite network configs.",
		CreateContext: resourceInstanceCreate,
		ReadContext:   resourceInstanceRead,
		DeleteContext: resourceInstanceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client

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
		UseJSONNumber: true,

		Schema: map[string]*schema.Schema{
			"allocate_params": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Description: "Nested argument with the constraints used to machine allocation. Defined below.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "The hostname of the MAAS machine to be allocated.",
						},
						"min_cpu_count": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							ForceNew:    true,
							Description: "The minimum number of cores used to allocate the MAAS machine.",
						},
						"min_memory": {
							Type:        schema.TypeInt,
							Optional:    true,
							Default:     0,
							ForceNew:    true,
							Description: "The minimum RAM memory size (in MB) used to allocate the MAAS machine.",
						},
						"pool": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "The pool name of the MAAS machine to be allocated.",
						},
						"system_id": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "The system_id of the MAAS machine to be allocated.",
						},
						"tags": {
							Type:        schema.TypeSet,
							Optional:    true,
							ForceNew:    true,
							Description: "A set of tag names that must be assigned on the MAAS machine to be allocated.",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"zone": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "The zone name of the MAAS machine to be allocated.",
						},
					},
				},
			},
			"cpu_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The number of CPU cores of the deployed MAAS machine.",
			},
			"deploy_params": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				MaxItems:    1,
				Description: "Nested argument with the config used to deploy the allocated machine. Defined below.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"distro_series": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "The distro series used to deploy the allocated MAAS machine. If it's not given, the MAAS server default value is used.",
						},
						"enable_hw_sync": {
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
							Description: "Periodically sync hardware",
						},
						"ephemeral": {
							Type:        schema.TypeBool,
							Optional:    true,
							ForceNew:    true,
							Description: "Deploy machine in memory",
						},
						"hwe_kernel": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "Hardware enablement kernel to use with the image. Only used when deploying Ubuntu.",
						},
						"user_data": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "Cloud-init user data script that gets run on the machine once it has deployed. A good practice is to set this with `file(\"/tmp/user-data.txt\")`, where `/tmp/user-data.txt` is a cloud-init script.",
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
			"ip_addresses": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of IP addressed assigned to the deployed MAAS machine.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"memory": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "The RAM memory size (in GiB) of the deployed MAAS machine.",
			},
			"network_interfaces": {
				Type:        schema.TypeSet,
				Optional:    true,
				ForceNew:    true,
				Description: "Specifies a network interface configuration done before the machine is deployed. Parameters defined below. This argument is processed in [attribute-as-blocks mode](https://www.terraform.io/docs/configuration/attr-as-blocks.html).",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip_address": {
							Type:             schema.TypeString,
							Optional:         true,
							ForceNew:         true,
							ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
							Description:      "Static IP address to be configured on the network interface. If this is set, the `subnet_cidr` is required.\n\n**NOTE:** If both `subnet_cidr` and `ip_address` are not defined, the interface will not be configured on the allocated machine.",
						},
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							ForceNew:    true,
							Description: "The name of the network interface to be configured on the allocated machine.",
						},
						"subnet_cidr": {
							Type:        schema.TypeString,
							Optional:    true,
							ForceNew:    true,
							Description: "An existing subnet CIDR used to configure the network interface. Unless `ip_address` is defined, a free IP address is allocated from the subnet.",
						},
					},
				},
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
			"zone": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The deployed MAAS machine zone name.",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

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
	_, err = waitForMachineStatus(ctx, client, machine.SystemID, []string{"Deploying"}, []string{"Deployed"}, d.Timeout(schema.TimeoutCreate))
	if err != nil {
		return diag.FromErr(err)
	}

	// Read MAAS machine info
	return resourceInstanceRead(ctx, d, meta)
}

func resourceInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

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

func resourceInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Release MAAS machine
	err := client.Machines.Release([]string{d.Id()}, "Released by Terraform")
	if err != nil {
		return diag.FromErr(err)
	}

	// Wait MAAS machine to be released
	_, err = waitForMachineStatus(ctx, client, d.Id(), []string{"Releasing"}, []string{"Ready"}, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getMachinesAllocateParams(d *schema.ResourceData) *entity.MachineAllocateParams {
	if p, ok := d.GetOk("allocate_params"); ok {
		allocateParamsData := p.([]interface{})
		if allocateParamsData[0] != nil {
			allocateParams := allocateParamsData[0].(map[string]interface{})
			return &entity.MachineAllocateParams{
				CPUCount: allocateParams["min_cpu_count"].(int),
				Mem:      int64(allocateParams["min_memory"].(int)),
				Name:     allocateParams["hostname"].(string),
				Zone:     allocateParams["zone"].(string),
				Pool:     allocateParams["pool"].(string),
				SystemID: allocateParams["system_id"].(string),
				Tags:     convertToStringSlice(allocateParams["tags"].(*schema.Set).List()),
			}
		}
	}
	return &entity.MachineAllocateParams{}
}

func getMachineDeployParams(d *schema.ResourceData) *entity.MachineDeployParams {
	if p, ok := d.GetOk("deploy_params"); ok {
		deployParamsData := p.([]interface{})
		if deployParamsData[0] != nil {
			deployParams := deployParamsData[0].(map[string]interface{})
			return &entity.MachineDeployParams{
				DistroSeries:    deployParams["distro_series"].(string),
				EnableHwSync:    deployParams["enable_hw_sync"].(bool),
				EphemeralDeploy: deployParams["ephemeral"].(bool),
				HWEKernel:       deployParams["hwe_kernel"].(string),
				UserData:        base64Encode([]byte(deployParams["user_data"].(string))),
			}
		}
	}
	return &entity.MachineDeployParams{}
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
