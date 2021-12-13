package maas

import (
	"fmt"
	"log"
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

func resourceMaasNetworkInterfaceBond() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceNetworkInterfaceBondCreate,
		ReadContext:   resourceNetworkInterfaceBondRead,
		UpdateContext: resourceNetworkInterfaceBondUpdate,
		DeleteContext: resourceNetworkInterfaceBondDelete,

		Schema: map[string]*schema.Schema{
			"machine": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"mac_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"vlan": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"parents": {
				Type:     schema.TypeList,
				Required: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type:  schema.TypeString,
				},
			},
			"bond_mode": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:		  true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
					"active-backup",
					"balance-rr",
					"balance-alb",
					"balance-tlb",
					"balance-xor",
					"broadcast",
					"802.3ad",
				}, false)),
			},
			"bond_lacp_rate": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:		  true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
					"fast",
					"slow",
				}, false)),
			},
			"bond_hash_policy": {
				Type:             schema.TypeString,
				Optional:         true,
				Computed:		  true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{
					"layer2",
					"layer2+3",
					"layer3+4",
					"encap2+3",
					"encap3+4",
				}, false)),
			},
			"mtu": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceNetworkInterfaceBondCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	bondParams := getNetworkInterfaceBondParams(d)

	bond, err := client.NetworkInterfaces.CreateBond(machine.SystemID, bondParams)

	// Save the resource id
	d.SetId(fmt.Sprintf("%v", bond.ID))

	return resourceNetworkInterfaceBondUpdate(ctx, d, m)
}

func resourceNetworkInterfaceBondRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Get params for the read operation
	bondID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] in bondRead: %s %d\n", machine.SystemID, bondID)

	// Get the network interface bond
	bond, err := getNetworkInterfaceBond(client, machine.SystemID, bondID)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[DEBUG] in bondRead: %#v\n", bond)

	// Set the Terraform state
	if err := d.Set("mac_address", bond.MACAddress); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNetworkInterfaceBondUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
//	client := m.(*client.Client)
//
//	// Get params for the update operation
//	linkID, err := strconv.Atoi(d.Id())
//	if err != nil {
//		return diag.FromErr(err)
//	}
//	machine, err := getMachine(client, d.Get("machine").(string))
//	if err != nil {
//		return diag.FromErr(err)
//	}
//	networkInterface, err := getNetworkInterface(client, machine.SystemID, d.Get("network_interface").(string))
//	if err != nil {
//		return diag.FromErr(err)
//	}
//
//	// Run update operation
//	if _, err := client.Machine.ClearDefaultGateways(machine.SystemID); err != nil {
//		return diag.FromErr(err)
//	}
//	if d.Get("default_gateway").(bool) {
//		if _, err := client.NetworkInterface.SetDefaultGateway(machine.SystemID, networkInterface.ID, linkID); err != nil {
//			return diag.FromErr(err)
//		}
//	}

	return nil// resourceNetworkInterfaceBondRead(ctx, d, m)
}

func resourceNetworkInterfaceBondDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Get params for the delete operation
	bondID, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	machine, err := getMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	// Delete the network interface link
	if err := deleteNetworkInterfaceBond(client, machine.SystemID, bondID); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getNetworkInterfaceBondParams(d *schema.ResourceData) *entity.NetworkInterfaceBondParams {
	parents := []int{}
	for _, p := range convertToStringSlice(d.Get("parents")) {
		parent, err := strconv.Atoi(p)
		if err != nil {
			panic(err)
		}
		parents = append(parents, parent)
	}

	return &entity.NetworkInterfaceBondParams{
		NetworkInterfacePhysicalParams: entity.NetworkInterfacePhysicalParams{
			MACAddress:		d.Get("mac_address").(string),
			Name:			d.Get("name").(string),
			VLAN:			d.Get("vlan").(string),
			MTU:			d.Get("mtu").(int),
		},
		Parents:         	 parents,
		BondMode:      		 d.Get("bond_mode").(string),
		BondLACPRate:  		 d.Get("bond_lacp_rate").(string),
		BondXMitHashPolicy:  d.Get("bond_hash_policy").(string),
	}
}

func getNetworkInterfaceBond(client *client.Client, machineSystemID string, networkInterfaceID int) (*entity.NetworkInterface, error) {
	networkInterface, err := client.NetworkInterface.Get(machineSystemID, networkInterfaceID)
	if err != nil {
		return nil, err
	}
	return networkInterface, nil
}

func deleteNetworkInterfaceBond(client *client.Client, machineSystemID string, networkInterfaceID int) error {
	err := client.NetworkInterface.Delete(machineSystemID, networkInterfaceID)
	return err
}
