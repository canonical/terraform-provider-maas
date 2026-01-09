package maas

import (
	"context"
	"fmt"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceMAASRackControllers() *schema.Resource {
	return &schema.Resource{
		Description: "Provides details about all MAAS rack controllers.",
		ReadContext: dataSourceMAASRackControllersRead,

		Schema: map[string]*schema.Schema{
			"controllers": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of all rack controllers.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"hostname": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The hostname of the rack controller.",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The system ID of the rack controller.",
						},
						"services": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of services running on the rack controller.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The name of the service.",
									},
									"status": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The status of the service.",
									},
								},
							},
						},
						"subnets": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of subnets accessible by the rack controller.",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"cidr": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The CIDR of the subnet.",
									},
									"fabric": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The fabric name of the subnet.",
									},
									"id": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "The ID of the subnet.",
									},
									"vlan": {
										Type:        schema.TypeInt,
										Computed:    true,
										Description: "The VLAN ID of the subnet.",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceMAASRackControllersRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// Get all rack controllers
	rackControllers, err := client.RackControllers.Get(&entity.RackControllersGetParams{})
	if err != nil {
		return diag.FromErr(err)
	}

	var controllers []any

	for _, rackController := range rackControllers {
		servicesList := make([]map[string]any, len(rackController.ServiceSet))
		for i, service := range rackController.ServiceSet {
			servicesList[i] = map[string]any{
				"name":   service.Name,
				"status": service.Status,
			}
		}

		// Extract unique subnets from all interface links
		subnetMap := make(map[int]map[string]any)

		for _, iface := range rackController.InterfaceSet {
			for _, link := range iface.Links {
				subnetID := link.Subnet.ID
				if _, exists := subnetMap[subnetID]; !exists {
					subnetMap[subnetID] = map[string]any{
						"fabric": link.Subnet.VLAN.Fabric,
						"cidr":   link.Subnet.CIDR,
						"vlan":   link.Subnet.VLAN.VID,
						"id":     fmt.Sprintf("%d", subnetID),
					}
				}
			}
		}

		// Convert map to slice
		subnetsList := make([]any, 0, len(subnetMap))
		for _, subnet := range subnetMap {
			subnetsList = append(subnetsList, subnet)
		}

		controller := map[string]any{
			"hostname": rackController.Hostname,
			"id":       rackController.SystemID,
			"services": servicesList,
			"subnets":  subnetsList,
		}

		controllers = append(controllers, controller)
	}

	d.SetId("maas_rack_controllers")

	if err := d.Set("controllers", controllers); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
