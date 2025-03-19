package maas

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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
				Computed:    true,
				Description: "The boot source database ID this resource set is associated with.",
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

	err := awaitImportComplete(client)
	if err != nil {
		return diag.Errorf("Could not await image importing: %v", err)
	}

	resources, err := getBootResources(client, "synced")
	if err != nil {
		return diag.Errorf("error fetching synced boot resources: %v", err)
	}
	resourceMap := make(map[string]struct{})
	for _, res := range resources {
		resourceMap[res.Name] = struct{}{}
	}

	bootselections := d.Get("boot_source_selections").(*schema.Set).List()
	if len(bootselections) == 0 {
		return diag.Errorf("At least one boot source selection must be added to the boot resources")
	}
	bootsource, err := getBootSource(client)
	if err != nil {
		return diag.Errorf("error fetching boot source: %v", err)
	}
	d.SetId(fmt.Sprintf("%v", bootsource.ID))

	for _, bootselection := range bootselections {
		bootselection, err := getBootSourceSelection(client, bootsource.ID, bootselection.(int))
		if err != nil {
			return diag.Errorf("error fetching boot selection %v: %s", bootselection, err)
		}

		// check the selection has it's resource created
		if _, exists := resourceMap[fmt.Sprintf("%s/%s", bootselection.OS, bootselection.Release)]; !exists {
			return diag.Errorf("Boot Resource missing for %s/%s", bootselection.OS, bootselection.Release)
		}
	}

	return resourceBootResourcesRead(ctx, d, meta)
}

func resourceBootResourcesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	err := awaitImportComplete(client)
	if err != nil {
		return diag.Errorf("Could not await image importing: %v", err)
	}

	resources, err := getBootResources(client, "synced")
	if err != nil {
		return diag.Errorf("error fetching synced boot resources: %v", err)
	}
	bootsource, err := getBootSource(client)
	if err != nil {
		return diag.Errorf("error fetching boot source: %v", err)
	}
	d.SetId(fmt.Sprintf("%v", bootsource.ID))

	selections := d.Get("boot_source_selections")
	if selections == nil {
		return diag.Errorf("boot_source_selection is missing from Resources state")
	}
	selectionMap := make(map[int]struct{})
	for _, sel := range selections.(*schema.Set).List() {
		if id, ok := sel.(int); ok {
			selectionMap[id] = struct{}{}
		} else {
			log.Printf("[DEBUG] Invalid selection ID found in state: %v", sel)
		}
	}

	// TODO: This seems unclean, is there a smarter way to get the selection IDs?
	var selectionSet []int
	for _, res := range resources {
		parts := strings.SplitN(res.Name, "/", 2)
		if len(parts) < 2 {
			return diag.Errorf("Invalid resource name: %s", res.Name)
		}
		os, release := parts[0], parts[1]

		// avoid the bootloaders
		if strings.Contains(os, "efi") || strings.Contains(os, "pxe") || strings.Contains(os, "grub") {
			continue
		}

		selection, err := getBootSourceSelectionByRelease(client, bootsource.ID, os, release)
		if err != nil {
			return diag.Errorf("error fetching boot selection '%v/%v': %v", os, release, err)
		}
		if selection == nil {
			log.Printf("[DEBUG] No selection found in MAAS for %s %s\n", os, release)
		}
		if _, exists := selectionMap[res.ID]; exists {
			selectionSet = append(selectionSet, selection.ID)
			log.Printf("[DEBUG] %s %s found in MAAS attached to resource\n", os, release)
		} else {
			log.Printf("[DEBUG] %s %s found in MAAS but not attached to resource\n", os, release)
		}
	}

	tfState := map[string]interface{}{
		"boot_source":            bootsource.ID,
		"boot_source_selections": unique(selectionSet),
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

	err := awaitImportComplete(client)
	if err != nil {
		return diag.Errorf("Could not await image importing: %v", err)
	}

	resources, err := getBootResources(client, "synced")
	if err != nil {
		return diag.Errorf("error fetching synced boot resources: %v", err)
	}
	resourceMap := make(map[string]struct{})
	for _, res := range resources {
		resourceMap[res.Name] = struct{}{}
	}
	bootsource, err := getBootSource(client)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(fmt.Sprintf("%v", bootsource.ID))

	selections := d.Get("boot_source_selections").([]int)
	for _, selection := range selections {
		bootselection, err := getBootSourceSelection(client, bootsource.ID, selection)
		if err != nil {
			return diag.Errorf("error fetching boot selection %v: %s", bootselection, err)
		}

		if _, exists := resourceMap[fmt.Sprintf("%s/%s", bootselection.OS, bootselection.Release)]; !exists {
			return diag.Errorf("Boot Resource missing for %s/%s", bootselection.OS, bootselection.Release)
		}
	}

	return resourceBootResourcesRead(ctx, d, meta)
}

func resourceBootResourcesDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// if you delete a resource in terraform, we ensure the selections contained are also deleted
	client := meta.(*ClientConfig).Client

	err := awaitImportComplete(client)
	if err != nil {
		return diag.Errorf("Could not await image importing: %v", err)
	}

	bootsource, err := getBootSource(client)
	if err != nil {
		return diag.Errorf("Could not fetch boot source: %v", err)
	}

	// delete the selections attached to this resource

	bootselections := d.Get("boot_source_selections").(*schema.Set).List()

	resourceMap := make(map[int]struct{})
	for _, bootselection := range bootselections {
		resourceMap[bootselection.(int)] = struct{}{}
	}
	for _, bootselection := range bootselections {
		err := client.BootSourceSelection.Delete(bootsource.ID, bootselection.(int))
		if err != nil {
			// 400 if the selection is the default
			if !strings.Contains(err.Error(), "operating system used in ephemeral environments") {
				continue
			}

			return diag.Errorf("Could not delete selection '%v': %v", bootselection.(int), err)
		}
	}
	err = awaitImportComplete(client)
	if err != nil {
		return diag.Errorf("Could not await image importing: %v", err)
	}

	resources, err := getBootResources(client, "synced")
	if err != nil {
		return diag.Errorf("error fetching synced boot resources: %v", err)
	}
	for _, resource := range resources {
		parts := strings.SplitN(resource.Name, "/", 2)
		if len(parts) < 2 {
			return diag.Errorf("Invalid resource name: %s", resource.Name)
		}
		os, release := parts[0], parts[1]

		// the selection should be deleted
		bootsourceselection, err := findBootSourceSelection(client, bootsource.ID, os, release)
		if err != nil {
			// 404 means the resource was deleted already
			if !strings.Contains(err.Error(), "404 Not Found") {
				continue
			}
			// anything else is an error
			return diag.Errorf("error finding selection '%v/%v': %v", os, release, err)
		}

		if _, exists := resourceMap[bootsourceselection.ID]; exists {
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

func awaitImportComplete(client *client.Client) error {
	if err := client.BootResources.Import(); err != nil {
		return err
	}
	timeout := 40 * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	result := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		if importing, err := client.BootResources.IsImporting(); err != nil {
			return retry.NonRetryableError(err)
		} else if importing {
			return retry.RetryableError(fmt.Errorf("boot resources still importing, waiting... "))
		}
		return nil
	})
	// wait for everything to take effect
	time.Sleep(10 * time.Second)
	return result
}

func unique(values []int) []int {
	// prepopulate the memory required to increase speed
	foundValues := make(map[int]struct{}, len(values))
	for _, v := range values {
		foundValues[v] = struct{}{}
	}

	// convert from mapping to slice, with pre-alloc of memory
	output := make([]int, len(foundValues))
	i := 0
	for k := range foundValues {
		output[i] = k
		i++
	}
	return output
}
