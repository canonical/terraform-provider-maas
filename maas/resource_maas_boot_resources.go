package maas

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceMAASBootResources() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS boot resources.",
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

func resourceBootResourcesCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	bootSelections := d.Get("boot_source_selections").(*schema.Set).List()

	if len(bootSelections) == 0 {
		return diag.Errorf("At least one boot source selection must be added to the boot resources")
	}

	bootSource, err := getBootSource(client)
	if err != nil {
		return diag.Errorf("error fetching boot source: %v", err)
	}

	d.Set("boot_source", bootSource.ID)
	d.SetId(fmt.Sprintf("%v", bootSource.ID))

	var selectionSet []int
	for _, bootSelection := range bootSelections {
		if slices.Contains(selectionSet, bootSelection.(int)) {
			log.Printf("[DEBUG] selection %v already discovered: %+v\n", bootSelection.(int), selectionSet)
		}

		bootSelection, err := getBootSourceSelection(client, bootSource.ID, bootSelection.(int))
		if err != nil {
			return diag.Errorf("error fetching boot selection %v: %s", bootSelection, err)
		}

		// check the selection has its resource created
		if _, exists := resourceMap[fmt.Sprintf("%s/%s", bootSelection.OS, bootSelection.Release)]; !exists {
			return diag.Errorf("Boot Resource missing for %s/%s", bootSelection.OS, bootSelection.Release)
		}

		selectionSet = append(selectionSet, bootSelection.ID)
	}

	if err := d.Set("boot_source_selections", unique(selectionSet)); err != nil {
		return diag.Errorf("Failed to set boot_source_selections: %v", err)
	}

	return resourceBootResourcesRead(ctx, d, meta)
}

func resourceBootResourcesRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	err := awaitImportComplete(client)
	if err != nil {
		return diag.Errorf("Could not await image importing: %v", err)
	}

	selections := d.Get("boot_source_selections").(*schema.Set).List()
	if selections == nil {
		return diag.Errorf("boot_source_selection is missing from Resources state")
	}

	selectionMap := make(map[int]struct{})

	for _, sel := range selections {
		if id, ok := sel.(int); ok {
			selectionMap[id] = struct{}{}
		} else {
			log.Printf("[DEBUG] Invalid selection ID found in state: %v", sel)
		}
	}

	resources, err := getBootResources(client, "synced")
	if err != nil {
		return diag.Errorf("error fetching synced boot resources: %v", err)
	}

	bootSource, err := getBootSource(client)
	if err != nil {
		return diag.Errorf("error fetching boot source: %v", err)
	}

	// TODO: This seems unclean, is there a smarter way to get the selection IDs?
	var selectionSet []int

	for _, res := range resources {
		parts := strings.SplitN(res.Name, "/", 2)
		if len(parts) < 2 {
			return diag.Errorf("Invalid resource name: %s", res.Name)
		}

		os, release := parts[0], parts[1]

		// avoid the boot loaders
		if strings.Contains(os, "efi") || strings.Contains(os, "pxe") || strings.Contains(os, "grub") {
			continue
		}

		selection, err := getBootSourceSelectionByRelease(client, bootSource.ID, os, release)
		if err != nil {
			return diag.Errorf("error fetching boot selection '%v/%v': %v", os, release, err)
		}

		if selection == nil {
			log.Printf("[DEBUG] No selection found in MAAS for %s %s\n", os, release)
			continue
		}

		if _, exists := selectionMap[selection.ID]; !exists {
			log.Printf("[DEBUG] %s %s found in MAAS but not attached to resource\n", os, release)
		} else {
			selectionSet = append(selectionSet, selection.ID)

			log.Printf("[DEBUG] %s %s found in MAAS attached to resource\n", os, release)
		}
	}

	tfState := map[string]any{
		"boot_source":            bootSource.ID,
		"boot_source_selections": unique(selectionSet),
	}

	d.SetId(fmt.Sprintf("%v", bootSource.ID))

	if err := setTerraformState(d, tfState); err != nil {
		return diag.Errorf("Could not apply terraform state: %v", err)
	}

	return nil
}

func resourceBootResourcesUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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

	bootSource, err := getBootSource(client)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", bootSource.ID))

	selections := d.Get("boot_source_selections").([]int)
	for _, selection := range selections {
		bootSourceSelection, err := getBootSourceSelection(client, bootSource.ID, selection)
		if err != nil {
			return diag.Errorf("error fetching boot selection %v: %s", bootSourceSelection, err)
		}

		if _, exists := resourceMap[fmt.Sprintf("%s/%s", bootSourceSelection.OS, bootSourceSelection.Release)]; !exists {
			return diag.Errorf("Boot Resource missing for %s/%s", bootSourceSelection.OS, bootSourceSelection.Release)
		}
	}

	return resourceBootResourcesRead(ctx, d, meta)
}

func resourceBootResourcesDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	// we trigger an image import to clean up any hanging resources
	err := awaitImportComplete(client)
	if err != nil {
		return diag.Errorf("Could not await image importing: %v", err)
	}

	d.SetId("")

	return nil
}

func getBootResources(client *client.Client, syncType string) ([]entity.BootResource, error) {
	// syncType: one of synched, uploaded
	readParams := entity.BootResourcesReadParams{
		Type: syncType,
	}

	bootResources, err := client.BootResources.Get(&readParams)
	if err != nil {
		return nil, err
	}

	return bootResources, nil
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
	// add a small delay to ensure the resources are fully updated
	if err := retry.RetryContext(ctx, 10*time.Second, func() *retry.RetryError {
		return nil
	}); err != nil {
		return fmt.Errorf("error after waiting 10 seconds: %s", err)
	}

	return result
}

func unique(values []int) []int {
	// pre-populate the memory required to increase speed
	foundValues := make(map[int]struct{}, len(values))
	for _, v := range values {
		foundValues[v] = struct{}{}
	}

	// convert from mapping to slice, with pre-alloc of memory
	output := make([]int, 0, len(foundValues))
	for k := range foundValues {
		output = append(output, k)
	}

	return output
}
