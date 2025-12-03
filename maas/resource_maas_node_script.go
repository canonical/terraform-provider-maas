package maas

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
		Description: "Provides a resource to manage MAAS node scripts. It expects a script with " +
			"metadata defined only embedded in the script, and it computes them in the Terraform " +
			"state after the resource creation. Details about script metadata can be found in MAAS " +
			"docs, ref: https://maas.io/docs/reference-commissioning-scripts",
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
				Computed:    true,
				Description: "Whether the provided network configuration is applied before the node script runs.",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A description of what the node script does.",
			},
			"destructive": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether or not the node script overwrites data on any drive on the running system.",
			},
			"for_hardware": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A list of modalias, PCI IDs, and/or USB IDs the node script will automatically run on.",
			},
			"hardware_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Defines what type of hardware the node script is associated with. May be `cpu`, `memory`, `storage`, `network`, `gpu`, or `node`.",
			},
			"may_reboot": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether or not the node script may reboot the system while running.",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the node script.",
			},
			"packages": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Dictionary of packages to be installed or extracted before running the node script.",
			},
			"parallel": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Whether the node script may be run in parallel with other scripts. May be `disabled` to run by itself, `instance` to run along scripts with the same name, or `any` to run along any script.",
			},
			"parameters": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The parameters the node script accepts.",
			},
			"recommission": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether builtin commissioning scripts should be rerun after successfully running this node script.",
			},
			"results": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The results the node script will return on completion.",
			},
			"script": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The node script content encoded in base64.",
			},
			"script_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Defines when the node script should be used: `commissioning` or `testing` or `release`.",
			},
			"tags": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "A set of tag names assigned to the node script.",
			},
			"timeout": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "How long the node script is allowed to run before failing. `0` means unlimited time. " +
					"The time is represented in the following format: `[DD] [[HH:]MM:]ss[.uuuuuu]`",
			},
			"title": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The title of the node script.",
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

	nodeScript, err := client.NodeScripts.Create(nil, scriptRaw)
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

	parametersJSON, err := json.Marshal(nodeScript.Parameters)
	if err != nil {
		return diag.FromErr(err)
	}

	resultsJSON, err := json.Marshal(nodeScript.Results)
	if err != nil {
		return diag.FromErr(err)
	}

	// Detect the latest change of the node script in the script history. A bigger database ID
	// means a newer version of the script. In the Terraform provider implementation we only care
	// about the most recent version of the script. When we find it, we keep track of its index in
	// the slice. We are using this index to set the proper script content in the Terraform state.
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
		"parameters":                  string(parametersJSON),
		"recommission":                nodeScript.Recommission,
		"results":                     string(resultsJSON),
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

	nodeScript, err := client.NodeScript.Update(d.Id(), nil, scriptRaw)
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
