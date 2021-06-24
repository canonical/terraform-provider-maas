package maas

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			"machine_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"network_interface_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"subnet_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"mode": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  "AUTO",
				ValidateDiagFunc: func(value interface{}, path cty.Path) diag.Diagnostics {
					v := value.(string)
					if !(v == "AUTO" || v == "DHCP" || v == "STATIC") {
						return diag.FromErr(fmt.Errorf("mode must be 'AUTO', 'DHCP', or 'STATIC' (got '%s')", v))
					}
					return nil
				},
			},
			"ip_address": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
				ValidateDiagFunc: func(value interface{}, path cty.Path) diag.Diagnostics {
					v := value.(string)
					if ip := net.ParseIP(v); ip == nil {
						return diag.FromErr(fmt.Errorf("ip_address must be a valid IP address (got '%s')", v))
					}
					return nil
				},
			},
			"default_gateway": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceNetworkInterfaceLinkCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Get params for the create operation
	machineId := d.Get("machine_id").(string)
	networkInterfaceId := d.Get("network_interface_id").(int)
	params := getNetworkInterfaceLinkParams(client, d)

	// Create network interface link
	link, err := createNetworkInterfaceLink(client, machineId, networkInterfaceId, params)
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
	linkId, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	machineId := d.Get("machine_id").(string)
	networkInterfaceId := d.Get("network_interface_id").(int)

	// Get the network interface link
	link, err := getNetworkInterfaceLink(client, machineId, networkInterfaceId, linkId)
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
	linkId, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	machineId := d.Get("machine_id").(string)
	networkInterfaceId := d.Get("network_interface_id").(int)
	params := getNetworkInterfaceLinkParams(client, d)

	// Run update operation
	_, err = client.Machine.ClearDefaultGateways(machineId)
	if err != nil {
		return diag.FromErr(err)
	}
	if params.DefaultGateway {
		_, err = client.NetworkInterface.SetDefaultGateway(machineId, networkInterfaceId, linkId)
		if err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceNetworkInterfaceLinkRead(ctx, d, m)
}

func resourceNetworkInterfaceLinkDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Get params for the delete operation
	linkId, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	machineId := d.Get("machine_id").(string)
	networkInterfaceId := d.Get("network_interface_id").(int)

	// Delete the network interface link
	err = deleteNetworkInterfaceLink(client, machineId, networkInterfaceId, linkId)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getNetworkInterfaceLinkParams(client *client.Client, d *schema.ResourceData) *entity.NetworkInterfaceLinkParams {
	params := entity.NetworkInterfaceLinkParams{
		Subnet:         d.Get("subnet_id").(int),
		Mode:           d.Get("mode").(string),
		DefaultGateway: d.Get("default_gateway").(bool),
	}

	if p, ok := d.GetOk("ip_address"); ok {
		params.IPAddress = p.(string)
	}

	return &params
}

func createNetworkInterfaceLink(client *client.Client, machineId string, networkInterfaceId int, params *entity.NetworkInterfaceLinkParams) (*entity.NetworkInterfaceLink, error) {
	// Clear existing links
	_, err := client.NetworkInterface.Disconnect(machineId, networkInterfaceId)
	if err != nil {
		return nil, err
	}

	// Create new link
	networkInterface, err := client.NetworkInterface.LinkSubnet(machineId, networkInterfaceId, params)
	if err != nil {
		return nil, err
	}

	return &networkInterface.Links[0], nil
}

func getNetworkInterfaceLink(client *client.Client, machineId string, networkInterfaceId int, linkId int) (*entity.NetworkInterfaceLink, error) {
	networkInterface, err := client.NetworkInterface.Get(machineId, networkInterfaceId)
	if err != nil {
		return nil, err
	}

	for _, link := range networkInterface.Links {
		if link.ID == linkId {
			return &link, nil
		}
	}

	return nil, fmt.Errorf("cannot find link with id '%v'", linkId)
}

func deleteNetworkInterfaceLink(client *client.Client, machineId string, networkInterfaceId int, linkId int) (err error) {
	_, err = client.NetworkInterface.UnlinkSubnet(machineId, networkInterfaceId, linkId)
	return
}
