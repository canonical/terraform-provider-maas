package maas

import (
	"context"
	"fmt"
	"net"

	"github.com/hashicorp/go-cty/cty"
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
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				client := m.(*client.Client)
				machine, err := findMachine(client, d.Id())
				if err != nil {
					return nil, err
				}
				if machine.StatusName != "Deployed" {
					return nil, fmt.Errorf("machine '%s' needs to be already deployed to be imported as maas_instance resource", machine.Hostname)
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"allocate_params": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_cpu_count": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"min_memory": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"hostname": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"zone": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"pool": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"tags": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
			"deploy_params": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"distro_series": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"hwe_kernel": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"user_data": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
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
						"subnet_cidr": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ip_address": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
							ValidateDiagFunc: func(value interface{}, path cty.Path) diag.Diagnostics {
								v := value.(string)
								if ip := net.ParseIP(v); ip == nil {
									return diag.FromErr(fmt.Errorf("ip_address must be a valid IP address (got '%s')", v))
								}
								return nil
							},
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
	p, ok := d.GetOk("network_interfaces")
	if !ok {
		return nil
	}
	machineNics, err := client.NetworkInterfaces.Get(machine.SystemID)
	if err != nil {
		return err
	}
	subnets, err := client.Subnets.Get()
	if err != nil {
		return err
	}
	nics := p.([]interface{})
	for _, nic := range nics {
		n := nic.(map[string]interface{})
		// Find the machine network interface
		name := n["name"].(string)
		var nicFound *entity.NetworkInterface = nil
		for _, machineNic := range machineNics {
			if machineNic.Name == name {
				nicFound = &machineNic
				break
			}
		}
		if nicFound == nil {
			return fmt.Errorf("network interface '%s' was not found on allocated instance '%s'", name, machine.FQDN)
		}
		subnetCidr := n["subnet_cidr"].(string)
		ipAddress := n["ip_address"].(string)
		if subnetCidr == "" {
			if ipAddress == "" {
				// Clear existing network interface links
				_, err := client.NetworkInterface.Disconnect(machine.SystemID, nicFound.ID)
				if err != nil {
					return err
				}
				continue
			} else {
				return fmt.Errorf("network interface '%s': the 'subnet_cidr' is required when 'ip_address' is set", name)
			}
		}
		// Find the subnet
		var subnetFound *entity.Subnet = nil
		for _, subnet := range subnets {
			if subnet.CIDR == subnetCidr {
				subnetFound = &subnet
				break
			}
		}
		if subnetFound == nil {
			return fmt.Errorf("subnet with CIDR '%s' was not found", subnetCidr)
		}
		// Prepare the network interface link parameters
		mode := "AUTO"
		if ipAddress != "" {
			mode = "STATIC"
		}
		params := entity.NetworkInterfaceLinkParams{
			Mode:      mode,
			Subnet:    subnetFound.ID,
			IPAddress: ipAddress,
		}
		// Clear existing network interface links
		_, err := client.NetworkInterface.Disconnect(machine.SystemID, nicFound.ID)
		if err != nil {
			return err
		}
		// Create new network interface link
		_, err = client.NetworkInterface.LinkSubnet(machine.SystemID, nicFound.ID, &params)
		if err != nil {
			return err
		}
	}
	return nil
}
