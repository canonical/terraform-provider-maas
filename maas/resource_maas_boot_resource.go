package maas

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceMAASBootResources() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS bootresources.",
		CreateContext: resourceBootResourcesCreate,
		ReadContext:   resourceBootResourcesRead,
		UpdateContext: resourceBootResourcesUpdate,
		DeleteContext: resourceBootResourcesDelete,

		Schema: map[string]*schema.Schema{
			"boot_source": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The boot source database ID this resource is associated with.",
			},
			"boot_source_selections": {
				Type:        schema.TypeSet,
				Required:    true,
				Description: "The set of database IDs for boot source selections to attach to this boot resource",
				Elem:        &schema.Schema{Type: schema.TypeInt},
			},
		},
	}
}

func resourceBootResourcesCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	for isImporting(client) {
		time.Sleep(5)
	}

	resources, err := getBootResources(client, "synced")
	if err != nil {
		return diag.FromErr(err)
	}
	resourceMap := make(map[string]struct{})
	for _, res := range resources {
		resourceMap[res.Name] = struct{}{}
	}

	bootselections := d.Get("boot_source_selections").(*schema.Set).List()
	if len(bootselections) == 0 {
		return diag.Errorf("At least one boot source selection must be added to the boot resources")
	}

	for _, bootselection := range bootselections {
		bootselection, err := getBootSourceSelection(client, d.Get("boot_source").(int), bootselection.(int))
		if err != nil {
			return diag.FromErr(err)
		}
		// check the selection has it's resource created
		if _, exists := resourceMap[fmt.Sprintf("%s/%s", bootselection.OS, bootselection.Release)]; !exists {
			return diag.Errorf("Boot Resource missing for %s/%s", bootselection.OS, bootselection.Release)
		}
	}

	return nil
}

func resourceBootResourcesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	resources, err := getBootResources(client, "synced")
	if err != nil {
		return diag.FromErr(err)
	}
	bootsource, err := getBootSource(client)
	if err != nil {
		return diag.FromErr(err)
	}

	// TODO: This seems unclean, is there a smarter way to get the selection IDs?
	var selectionSet []int
	for _, res := range resources {
		parts := strings.SplitN(res.Name, "/", 2)
		if len(parts) < 2 {
			return diag.Errorf("Invalid resource name: %s", res.Name)
		}
		os, release := parts[0], parts[1]

		selection, err := getBootSourceSelectionByRelease(client, bootsource.ID, os, release)
		if err != nil {
			return diag.FromErr(err)
		}
		selectionSet = append(selectionSet, selection.ID)
	}

	tfState := map[string]interface{}{
		"boot_source":            bootsource.ID,
		"boot_source_selections": selectionSet,
	}

	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBootResourcesUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Ensure newly created selections exist
	// TODO: How to see if old selections no longer exist
	client := meta.(*ClientConfig).Client

	for isImporting(client) {
		time.Sleep(5)
	}

	return resourceBootResourcesRead(ctx, d, meta)
}

func resourceBootResourcesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// if you delete a resource in terraform, we ensure the selections contained are also deleted
	client := meta.(*ClientConfig).Client

	for isImporting(client) {
		time.Sleep(time.Second * 5)
	}

	existing, err := getBootResources(client, "synced")
	if err != nil {
		return diag.FromErr(err)
	}
	for _, resource := range existing {
		parts := strings.SplitN(resource.Name, "/", 2)
		if len(parts) < 2 {
			return diag.Errorf("Invalid resource name: %s", resource.Name)
		}
		os, release := parts[0], parts[1]

		// the selection should be deleted
		bootsourceselection, err := findBootSourceSelection(client, d.Get("boot_source").(int), os, release)
		if err != nil {
			return diag.FromErr(err)
		}
		if bootsourceselection != nil {
			return diag.Errorf("boot source selection (%s %s) was unexpectedly found on deleted resource", os, release)
		}
	}

	return resourceBootResourcesRead(ctx, d, meta)
}

func getBootResources(client *client.Client, synctype string) ([]entity.BootResource, error) {
	// synctype: one of synched, uploaded
	readparams := entity.BootResourcesReadParams{
		Type: synctype,
	}
	bootresources, err := client.BootResources.Get(&readparams)
	if err != nil {
		return nil, err
	}
	return bootresources, nil
}

func isImporting(client *client.Client) bool {
	importing, _ := client.BootResources.IsImporting()
	return importing
}
