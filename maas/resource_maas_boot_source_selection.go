package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/go-set/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceMAASBootSourceSelection() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage a MAAS boot source selection.",
		CreateContext: resourceBootSourceSelectionCreate,
		ReadContext:   resourceBootSourceSelectionRead,
		UpdateContext: resourceBootSourceSelectionUpdate,
		DeleteContext: resourceBootSourceSelectionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceBootSourceSelectionImport,
		},

		Schema: map[string]*schema.Schema{
			"arches": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringInSlice(
						[]string{"amd64", "arm64", "armhf", "i386", "ppc64el", "s390x"},
						false,
					),
				},
				Required:    true,
				Description: "The architecture list for this selection. Valid architectures are: `[\"amd64\", \"arm64\", \"armhf\", \"i386\", \"ppc64el\", \"s390x\"]`",
			},
			"boot_source": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "The boot source database ID this selection is associated with.",
			},
			"labels": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
				Description: "The label list for this selection. Default is: `[\"*\"]`",
			},
			"os": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The operating system for this selection.",
			},
			"release": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The specific release of the operating system for this selection.",
			},
			"subarches": {
				Type:        schema.TypeSet,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Optional:    true,
				Computed:    true,
				Description: "The list of subarches for this selection. Default is: `[\"*\"]`",
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Minute),
			Update: schema.DefaultTimeout(30 * time.Minute),
			Delete: schema.DefaultTimeout(30 * time.Minute),
		},
	}
}

func resourceBootSourceSelectionImport(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), ":")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected BOOT_SOURCE:BOOT_SOURCE_SELECTION_ID", d.Id())
	}

	bootSourceID, err := strconv.Atoi(idParts[0])
	if err != nil {
		return nil, fmt.Errorf("failed to parse BOOT_SOURCE id, expected int-like, err: %v", err)
	}

	d.Set("boot_source", bootSourceID)
	d.SetId(idParts[1])

	return []*schema.ResourceData{d}, nil
}

func resourceBootSourceSelectionCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	arches := convertToStringSlice(d.Get("arches").(*schema.Set).List())

	subarches := convertToStringSlice(d.Get("subarches").(*schema.Set).List())
	if len(subarches) == 0 {
		subarches = append(subarches, "*")
	}

	labels := convertToStringSlice(d.Get("labels").(*schema.Set).List())
	if len(labels) == 0 {
		labels = append(labels, "*")
	}

	bootSourceSelectionParams := entity.BootSourceSelectionParams{
		OS:        d.Get("os").(string),
		Release:   d.Get("release").(string),
		Arches:    arches,
		Subarches: subarches,
		Labels:    labels,
	}

	// Create the selection
	bootSourceSelection, err := client.BootSourceSelections.Create(d.Get("boot_source").(int), &bootSourceSelectionParams)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating %s %s: %s", d.Get("os"), d.Get("release"), err))
	}

	// Trigger image import and wait for its completion
	if err := awaitImportComplete(client, d.Get("os").(string), d.Get("release").(string), arches, d.Timeout(schema.TimeoutCreate)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", bootSourceSelection.ID))

	return resourceBootSourceSelectionRead(ctx, d, meta)
}

func resourceBootSourceSelectionRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	bootSourceSelection, err := getBootSourceSelection(client, d.Get("boot_source").(int), id)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%v", bootSourceSelection.ID))

	tfState := map[string]any{
		"arches":      bootSourceSelection.Arches,
		"boot_source": bootSourceSelection.BootSourceID,
		"labels":      bootSourceSelection.Labels,
		"os":          bootSourceSelection.OS,
		"release":     bootSourceSelection.Release,
		"subarches":   bootSourceSelection.Subarches,
	}

	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBootSourceSelectionUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	arches := convertToStringSlice(d.Get("arches").(*schema.Set).List())

	subarches := convertToStringSlice(d.Get("subarches").(*schema.Set).List())
	if len(subarches) == 0 {
		subarches = append(subarches, "*")
	}

	labels := convertToStringSlice(d.Get("labels").(*schema.Set).List())
	if len(labels) == 0 {
		labels = append(labels, "*")
	}

	bootSourceSelectionParams := entity.BootSourceSelectionParams{
		OS:        d.Get("os").(string),
		Release:   d.Get("release").(string),
		Arches:    arches,
		Subarches: subarches,
		Labels:    labels,
	}

	// Update the selection
	if _, err := client.BootSourceSelection.Update(d.Get("boot_source").(int), id, &bootSourceSelectionParams); err != nil {
		return diag.FromErr(err)
	}

	// Trigger image import and wait for its completion
	if err := awaitImportComplete(client, d.Get("os").(string), d.Get("release").(string), arches, d.Timeout(schema.TimeoutUpdate)); err != nil {
		return diag.FromErr(err)
	}

	return resourceBootSourceSelectionRead(ctx, d, meta)
}

func resourceBootSourceSelectionDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}

	// Delete the selection
	if err := client.BootSourceSelection.Delete(d.Get("boot_source").(int), id); err != nil {
		// 404 means the resource was deleted already
		if strings.Contains(err.Error(), "404 Not Found") {
			return nil
		}

		return diag.FromErr(err)
	}

	// Trigger image import and wait for its completion to ensure image deletion
	if err := awaitImageDeleteComplete(client, d.Get("os").(string), d.Get("release").(string), d.Timeout(schema.TimeoutDelete)); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getBootSourceSelection(client *client.Client, bootSource int, id int) (*entity.BootSourceSelection, error) {
	bootSourceSelection, err := client.BootSourceSelection.Get(bootSource, id)
	if err != nil {
		return nil, err
	}

	if bootSourceSelection == nil {
		return nil, fmt.Errorf("boot source selection (%v %v) was not found", bootSource, id)
	}

	return bootSourceSelection, nil
}

func awaitImportComplete(client *client.Client, os string, release string, arches []string, timeout time.Duration) error {
	if err := client.BootResources.Import(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		// Retrieve all the boot resources with type synced. These resources are created when importing from a boot source selection.
		allResources, err := getBootResources(client, "synced")
		if err != nil {
			return retry.NonRetryableError(err)
		}

		// Confirm that for all user provided architectures in the boot source selection there is at least one boot resource found.
		// If any architecture is missing, then the import is still ongoing. To perform this operation, the usage of sets is required
		// since users can provide duplicates and MAAS can accept them.
		archesSet := set.From(arches)
		archesFound := set.New[string](0)

		for _, resource := range allResources {
			if resource.Name == fmt.Sprintf("%s/%s", os, release) {
				archesFound.Insert(strings.Split(resource.Architecture, "/")[0])

				// To verify that all boot resources are fully synced, we need to confirm the `completed` status of all their
				// individual files. This information is not accessible by the list read operation. Instead, we have to fetch
				// the resources individually.
				resourceDetails, err := client.BootResource.Get(resource.ID)
				if err != nil {
					return retry.NonRetryableError(err)
				}

				// For each boot resource the response contains a list of sets, with each set representing a version of the boot
				// resource. e.g., 24.04-ga-24.04-20250424. Each set has a complete boolean flag which, if set, represents the
				// completion of the image synchronization. If not, that means that the import is still ongoing.
				for _, resourceSet := range resourceDetails.Sets {
					if !resourceSet.Complete {
						return retry.RetryableError(fmt.Errorf("image still importing, waiting... "))
					}
				}
			}
		}

		// There is still difference between user selected architectures set and unique boot resource architectures for the given
		// os/release. The import is still ongoing.
		if !archesSet.Equal(archesFound) {
			return retry.RetryableError(fmt.Errorf("image still importing, waiting... "))
		}

		return nil
	})

	return result
}

func awaitImageDeleteComplete(client *client.Client, os string, release string, timeout time.Duration) error {
	if err := client.BootResources.Import(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	result := retry.RetryContext(ctx, timeout, func() *retry.RetryError {
		// Retrieve all the boot resources with type synced. These resources are created when importing from a boot source selection.
		allResources, err := getBootResources(client, "synced")
		if err != nil {
			return retry.NonRetryableError(err)
		}

		// For the given os/release architecture, there are existing boot resources. The deletion is still ongoing.
		for _, resource := range allResources {
			if resource.Name == fmt.Sprintf("%s/%s", os, release) {
				return retry.RetryableError(fmt.Errorf("image still deleting, waiting... "))
			}
		}

		return nil
	})

	return result
}
