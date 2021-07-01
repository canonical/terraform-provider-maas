package maas

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

func resourceMaasNetworkInterfaceLink() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkInterfaceLinkCreate,
		ReadContext:   resourceNetworkInterfaceLinkRead,
		UpdateContext: resourceNetworkInterfaceLinkUpdate,
		DeleteContext: resourceNetworkInterfaceLinkDelete,

		Schema: map[string]*schema.Schema{
			"machine": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_interface": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mode": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Default:          "AUTO",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"AUTO", "DHCP", "STATIC", "LINK_UP"}, false)),
			},
			"default_gateway": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ip_address": {
				Type:             schema.TypeString,
				Optional:         true,
				ForceNew:         true,
				Computed:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.IsIPAddress),
			},
		},
	}
}

func resourceNetworkInterfaceLinkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

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
	link, err := createNetworkInterfaceLink(client, machine.SystemID, networkInterface.ID, getNetworkInterfaceLinkParams(d, subnet.ID))
	if err != nil {
		return diag.FromErr(err)
	}

	// Save the resource id
	d.SetId(fmt.Sprintf("%v", link.ID))

	return resourceNetworkInterfaceLinkUpdate(ctx, d, m)
}

func resourceNetworkInterfaceLinkRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

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

func resourceNetworkInterfaceLinkUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

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

	return resourceNetworkInterfaceLinkRead(ctx, d, m)
}

func resourceNetworkInterfaceLinkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

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

func createNetworkInterfaceLink(client *client.Client, machineSystemID string, networkInterfaceID int, params *entity.NetworkInterfaceLinkParams) (*entity.NetworkInterfaceLink, error) {
	// Clear existing links
	_, err := client.NetworkInterface.Disconnect(machineSystemID, networkInterfaceID)
	if err != nil {
		return nil, err
	}
	// Create new link
	networkInterface, err := client.NetworkInterface.LinkSubnet(machineSystemID, networkInterfaceID, params)
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
