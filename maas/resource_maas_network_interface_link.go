package maas

import (
	"context"
	"fmt"
	"strconv"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceMaasNetworkInterfaceLink() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage network configuration on a network interface.",
		CreateContext: resourceNetworkInterfaceLinkCreate,
		ReadContext:   resourceNetworkInterfaceLinkRead,
		UpdateContext: resourceNetworkInterfaceLinkUpdate,
		DeleteContext: resourceNetworkInterfaceLinkDelete,

		Schema: map[string]*schema.Schema{
			"default_gateway": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Boolean value. When enabled, it sets the subnet gateway IP address as the default gateway for the machine the interface belongs to. This option can only be used with the `AUTO` and `STATIC` modes. Defaults to `false`.",
			},
			"ip_address": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
				Description:      "Valid IP address (from the given subnet) to be configured on the network interface. Only used when `mode` is set to `STATIC`.",
			},
			"machine": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The identifier (system ID, hostname, or FQDN) of the machine with the network interface.",
			},
			"mode": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          "AUTO",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"AUTO", "DHCP", "STATIC", "LINK_UP"}, false)),
				Description:      "Connection mode to subnet. It defaults to `AUTO`. Valid options are:\n\t* `AUTO` - Random static IP address from the subnet.\n\t* `DHCP` - IP address from the DHCP on the given subnet.\n\t* `STATIC` - Use `ip_address` as static IP address.\n\t* `LINK_UP` - Bring the interface up only on the given subnet. No IP address will be assigned.",
			},
			"network_interface": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The identifier (MAC address, name, or ID) of the network interface.",
			},
			"subnet": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The identifier (CIDR or ID) of the subnet to be connected.",
			},
		},
	}
}

func resourceNetworkInterfaceLinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Create network interface link
	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	networkInterface, err := getNetworkInterface(client, machine.SystemID, d.Get("network_interface").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	subnet, err := getSubnet(client, d.Get("subnet").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	link, err := createNetworkInterfaceLink(client, machine.SystemID, networkInterface, getNetworkInterfaceLinkParams(d, subnet.ID))
	if err != nil {
		return diag.FromErr(err)
	}

	// Save the resource id
	d.SetId(fmt.Sprintf("%v", link.ID))

	return resourceNetworkInterfaceLinkRead(ctx, d, meta)
}

func resourceNetworkInterfaceLinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Get params for the read operation
	linkID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	networkInterface, err := getNetworkInterface(client, machine.SystemID, d.Get("network_interface").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the network interface link
	link, err := getNetworkInterfaceLink(client, machine.SystemID, networkInterface.ID, linkID)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the Terraform state
	if err := d.Set("ip_address", link.IPAddress); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkInterfaceLinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Get params for the update operation
	linkID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	networkInterface, err := getNetworkInterface(client, machine.SystemID, d.Get("network_interface").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// Run update operation
	if _, err := client.Machine.ClearDefaultGateways(machine.SystemID); err != nil {
		return diag.FromErr(err)
	}
	if d.Get("default_gateway").(bool) {
		if _, err := client.NetworkInterface.SetDefaultGateway(machine.SystemID, networkInterface.ID, linkID); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceNetworkInterfaceLinkRead(ctx, d, meta)
}

func resourceNetworkInterfaceLinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Get params for the delete operation
	linkID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	networkInterface, err := getNetworkInterface(client, machine.SystemID, d.Get("network_interface").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// Delete the network interface link
	if err := deleteNetworkInterfaceLink(client, machine.SystemID, networkInterface.ID, linkID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getNetworkInterfaceLinkParams(d *schema.ResourceData, subnetID int) *entity.NetworkInterfaceLinkParams {
	return &entity.NetworkInterfaceLinkParams{
		Subnet:         subnetID,
		Mode:           d.Get("mode").(string),
		DefaultGateway: d.Get("default_gateway").(bool),
		IPAddress:      d.Get("ip_address").(string),
	}
}

func createNetworkInterfaceLink(client *client.Client, machineSystemID string, networkInterface *entity.NetworkInterface, params *entity.NetworkInterfaceLinkParams) (*entity.NetworkInterfaceLink, error) {
	// Clear existing links
	// VLAN type interfaces are excluded since this action is not allowed by MAAS itself:
	// <https://github.com/canonical/maas/blob/master/src/maasserver/models/interface.py#L2001-L2006>
	if networkInterface.Type != "vlan" {
		_, err := client.NetworkInterface.Disconnect(machineSystemID, networkInterface.ID)
		if err != nil {
			return nil, err
		}
	}

	// Create new link
	networkInterface, err := client.NetworkInterface.LinkSubnet(machineSystemID, networkInterface.ID, params)
	if err != nil {
		return nil, err
	}
	return &networkInterface.Links[0], nil
}

func getNetworkInterfaceLink(client *client.Client, machineSystemID string, networkInterfaceID int, linkID int) (*entity.NetworkInterfaceLink, error) {
	networkInterface, err := client.NetworkInterface.Get(machineSystemID, networkInterfaceID)
	if err != nil {
		return nil, err
	}
	for _, link := range networkInterface.Links {
		if link.ID == linkID {
			return &link, nil
		}
	}
	return nil, fmt.Errorf("cannot find link (%v) on the network interface (%v) from machine (%s)", linkID, networkInterfaceID, machineSystemID)
}

func deleteNetworkInterfaceLink(client *client.Client, machineSystemID string, networkInterfaceID int, linkID int) error {
	_, err := client.NetworkInterface.UnlinkSubnet(machineSystemID, networkInterfaceID, linkID)
	return err
}
