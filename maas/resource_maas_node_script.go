package maas

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var hardwareTypeEnumToName = map[entity.NodeScriptHardwareType]string{
	entity.ScriptHardwareTypeNode:    "node",
	entity.ScriptHardwareTypeCPU:     "cpu",
	entity.ScriptHardwareTypeMemory:  "memory",
	entity.ScriptHardwareTypeStorage: "storage",
	entity.ScriptHardwareTypeNetwork: "network",
	entity.ScriptHardwareTypeGPU:     "gpu",
}

var scriptTypeEnumToName = map[entity.NodeScriptType]string{
	entity.ScriptTypeCommissioning: "commissioning",
	entity.ScriptTypeTesting:       "testing",
	entity.ScriptTypeRelease:       "release",
}

var parallelEnumToName = map[entity.NodeScriptParallel]string{
	entity.ScriptParallelDisabled: "disabled",
	entity.ScriptParallelInstance: "instance",
	entity.ScriptParallelAny:      "any",
}

func resourceMAASNodeScript() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS node scripts.",
		CreateContext: resourceNodeScriptCreate,
		ReadContext:   resourceNodeScriptRead,
		UpdateContext: resourceNodeScriptUpdate,
		DeleteContext: resourceNodeScriptDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta any) ([]*schema.ResourceData, error) {
				client := meta.(*ClientConfig).Client

				nodeScript, err := getNodeScript(client, d.Id())
				if err != nil {
					return nil, err
				}
				d.SetId(nodeScript.Name)
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"apply_configured_networking": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether to apply the provided network configuration before the script runs.",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "A description of what the script does.",
			},
			"destructive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether or not the script overwrites data on any drive on the running system. Destructive scripts can not be run on deployed systems. Defaults to `false`.",
			},
			"for_hardware": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
					ValidateFunc: validation.StringMatch(
						regexp.MustCompile(`^(modalias:|pci:|usb:|system_vendor:|system_product:|system_version:|mainboard_vendor:|mainboard_product:)`),
						"must start with 'modalias:', 'pci:', 'usb:', 'system_vendor:', 'system_product:', 'system_version:', 'mainboard_vendor:', or 'mainboard_product:'",
					),
				},
				Description: "A list of modalias, PCI IDs, and/or USB IDs the script will automatically run on. Must start with `modalias:`, `pci:`, `usb:`, `system_vendor:`, `system_product:`, `system_version:`, `mainboard_vendor:`, or `mainboard_product:`.",
			},
			"hardware_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"cpu", "memory", "storage", "network", "node", "gpu"}, false),
				Description:  "The hardware_type defines what type of hardware the script is associated with. May be `cpu`, `memory`, `storage`, `network`, `gpu`, or `node`.",
			},
			"may_reboot": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether or not the script may reboot the system while running.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the new node script.",
			},
			"packages": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsJSON,
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					oldMap, err := structure.ExpandJsonFromString(oldValue)
					if err != nil {
						return false
					}
					newMap, err := structure.ExpandJsonFromString(newValue)
					if err != nil {
						return false
					}
					return reflect.DeepEqual(oldMap, newMap)
				},
				StateFunc: func(v any) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				Description: "Dictionary of packages to be installed or extracted before running the script.",
			},
			"parallel": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"disabled", "instance", "any"}, false),
				Description:  "Whether the script may be run in parallel with other scripts. May be `disabled` to run by itself, `instance` to run along scripts with the same name, or `any` to run along any script.",
			},
			"recommission": {
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Whether builtin commissioning scripts should be rerun after successfully running this script.",
			},
			"script": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The node script content encoded in base64.",
			},
			"script_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringInSlice([]string{"commissioning", "testing", "release"}, false),
				Description:  "The script_type defines when the script should be used: `commissioning` or `testing` or `release`. Defaults to `testing`.",
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A set of tag names assigned to the new node script. This argument is computed if it's not given.",
			},
			"timeout": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "How long the script is allowed to run before failing. `0` gives unlimited time, defaults to `0`. Expects data in the format `DD HH:MM:SS.uuuuuu`, `DD HH:MM:SS,uuuuuu`, or as specified by ISO 8601 (e.g. `P4DT1H15M20S` which is equivalent to 4 1:15:20) or PostgreSQL’s day-time interval format (e.g. `3 days 04:05:06`).",
			},
			"title": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The title of the script.",
			},
		},
	}
}

func resourceNodeScriptCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	scriptContent := d.Get("script").(string)

	scriptRaw, err := base64.StdEncoding.DecodeString(scriptContent)
	if err != nil {
		return diag.FromErr(err)
	}

	nodeScript, err := client.NodeScripts.Create(getNodeScriptParams(d), []byte(scriptRaw))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(nodeScript.Name)

	return resourceNodeScriptRead(ctx, d, meta)
}

func resourceNodeScriptRead(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	nodeScript, err := client.NodeScript.Get(d.Id(), true)
	if err != nil {
		return diag.FromErr(err)
	}

	packagesJSON, err := json.Marshal(nodeScript.Packages)
	if err != nil {
		return diag.FromErr(err)
	}

	latestChangeIdx := -1
	latestChange := -1

	for i, h := range nodeScript.History {
		if h.ID > latestChange {
			latestChange = h.ID
			latestChangeIdx = i
		}
	}

	tfState := map[string]any{
		"apply_configured_networking": nodeScript.ApplyConfiguredNetworking,
		"description":                 nodeScript.Description,
		"destructive":                 nodeScript.Destructive,
		"for_hardware":                nodeScript.ForHardware,
		"hardware_type":               hardwareTypeEnumToName[nodeScript.HardwareType],
		"may_reboot":                  nodeScript.MayReboot,
		"name":                        nodeScript.Name,
		"packages":                    string(packagesJSON),
		"parallel":                    parallelEnumToName[nodeScript.Parallel],
		"recommission":                nodeScript.Recommission,
		"script":                      nodeScript.History[latestChangeIdx].Data,
		"script_type":                 scriptTypeEnumToName[nodeScript.Type],
		"tags":                        nodeScript.Tags,
		"timeout":                     nodeScript.Timeout,
		"title":                       nodeScript.Title,
	}
	if err := setTerraformState(d, tfState); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceNodeScriptUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	scriptContent := d.Get("script").(string)

	scriptRaw, err := base64.StdEncoding.DecodeString(scriptContent)
	if err != nil {
		return diag.FromErr(err)
	}

	nodeScript, err := client.NodeScript.Update(d.Id(), getNodeScriptParams(d), []byte(scriptRaw))
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(nodeScript.Name)

	return resourceNodeScriptRead(ctx, d, meta)
}

func resourceNodeScriptDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

	if err := client.NodeScript.Delete(d.Id()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getNodeScriptParams(d *schema.ResourceData) *entity.NodeScriptParams {
	nodeScriptParams := entity.NodeScriptParams{
		ApplyConfiguredNetworking: d.Get("apply_configured_networking").(bool),
		Description:               d.Get("description").(string),
		Destructive:               d.Get("destructive").(bool),
		ForHardware:               strings.Join(convertToStringSlice(d.Get("for_hardware").(*schema.Set).List()), ","),
		HardwareType:              d.Get("hardware_type").(string),
		MayReboot:                 d.Get("may_reboot").(bool),
		Name:                      d.Get("name").(string),
		Packages:                  d.Get("packages").(string),
		Parallel:                  d.Get("parallel").(string),
		Recommission:              d.Get("recommission").(bool),
		ScriptType:                d.Get("script_type").(string),
		Tags:                      strings.Join(convertToStringSlice(d.Get("tags").(*schema.Set).List()), ","),
		Timeout:                   d.Get("timeout").(string),
		Title:                     d.Get("title").(string),
	}

	return &nodeScriptParams
}

func findNodeScript(client *client.Client, identifier string) (*entity.NodeScript, error) {
	nodeScripts, err := client.NodeScripts.Get(&entity.NodeScriptReadParams{})
	if err != nil {
		return nil, err
	}

	for _, s := range nodeScripts {
		if fmt.Sprintf("%v", s.ID) == identifier || s.Name == identifier {
			return &s, nil
		}
	}

	return nil, err
}

func getNodeScript(client *client.Client, identifier string) (*entity.NodeScript, error) {
	nodeScript, err := findNodeScript(client, identifier)
	if err != nil {
		return nil, err
	}

	if nodeScript == nil {
		return nil, fmt.Errorf("node script (%s) was not found", identifier)
	}

	return nodeScript, nil
}
