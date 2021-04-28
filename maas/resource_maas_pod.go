package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ionutbalutoiu/gomaasclient/api/endpoint"
	"github.com/ionutbalutoiu/gomaasclient/gmaw"
	"github.com/ionutbalutoiu/gomaasclient/maas"
	"github.com/juju/gomaasapi"
)

func resourceMaasPod() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePodCreate,
		ReadContext:   resourcePodRead,
		UpdateContext: resourcePodUpdate,
		DeleteContext: resourcePodDelete,

		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"power_address": {
				Type:     schema.TypeString,
				Required: true,
			},
			"power_user": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"power_pass": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"zone": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"pool": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"cpu_over_commit_ratio": {
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  1.0,
			},
			"memory_over_commit_ratio": {
				Type:     schema.TypeFloat,
				Optional: true,
				Default:  1.0,
			},
			"default_macvlan_mode": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"resources_cores_available": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"resources_memory_available": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"resources_local_storage_available": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"resources_cores_total": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"resources_memory_total": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"resources_local_storage_total": {
				Type:     schema.TypeInt,
				Computed: true,
			},
		},
	}
}

func resourcePodCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*gomaasapi.MAASObject)

	// Create Pod
	podsManager := maas.NewPodsManager(gmaw.NewPods(client))
	pod, err := podsManager.Create(getPodCreateParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	// Save Id
	d.SetId(fmt.Sprintf("%v", pod.ID))

	// Return updated pod
	return resourcePodUpdate(ctx, d, m)
}

func resourcePodRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*gomaasapi.MAASObject)

	// Get Pod details
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	podManager, err := maas.NewPodManager(id, gmaw.NewPod(client))
	if err != nil {
		return diag.FromErr(err)
	}
	pod := podManager.Current()

	// Set Terraform state
	if err := d.Set("name", pod.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("zone", pod.Zone.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pool", pod.Pool.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", pod.Tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_cores_available", pod.Available.Cores); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_cores_total", pod.Total.Cores); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_memory_available", pod.Available.Memory); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_memory_total", pod.Total.Memory); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_local_storage_available", pod.Available.LocalStorage); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_local_storage_total", pod.Total.LocalStorage); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePodUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*gomaasapi.MAASObject)

	// Get the pod manager
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	podManager, err := maas.NewPodManager(id, gmaw.NewPod(client))
	if err != nil {
		return diag.FromErr(err)
	}

	// Update Pod options
	_, err = podManager.Update(getPodUpdateParams(d, podManager.Current()))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourcePodRead(ctx, d, m)
}

func resourcePodDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*gomaasapi.MAASObject)

	// Delete Pod
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	podManager, err := maas.NewPodManager(id, gmaw.NewPod(client))
	if err != nil {
		return diag.FromErr(err)
	}
	err = podManager.Delete()
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getPodCreateParams(d *schema.ResourceData) *endpoint.PodParams {
	params := endpoint.PodParams{
		Type:                  d.Get("type").(string),
		PowerAddress:          d.Get("power_address").(string),
		CPUOverCommitRatio:    d.Get("cpu_over_commit_ratio").(float64),
		MemoryOverCommitRatio: d.Get("memory_over_commit_ratio").(float64),
	}

	if p, ok := d.GetOk("power_user"); ok {
		params.PowerUser = p.(string)
	}
	if p, ok := d.GetOk("power_pass"); ok {
		params.PowerPass = p.(string)
	}

	return &params
}

func getPodUpdateParams(d *schema.ResourceData, pod *endpoint.Pod) *endpoint.PodParams {
	params := endpoint.PodParams{
		Type:                  pod.Type,
		Name:                  pod.Name,
		PowerAddress:          d.Get("power_address").(string),
		CPUOverCommitRatio:    pod.CPUOverCommitRatio,
		MemoryOverCommitRatio: pod.MemoryOverCommitRatio,
		DefaultMacvlanMode:    pod.DefaultMACVLANMode,
		Zone:                  pod.Zone.Name,
		Pool:                  pod.Pool.Name,
		Tags:                  strings.Join(pod.Tags, ","),
	}

	if p, ok := d.GetOk("power_pass"); ok {
		params.PowerPass = p.(string)
	}
	if p, ok := d.GetOk("name"); ok {
		params.Name = p.(string)
	}
	if p, ok := d.GetOk("zone"); ok {
		params.Zone = p.(string)
	}
	if p, ok := d.GetOk("pool"); ok {
		params.Pool = p.(string)
	}
	if p, ok := d.GetOk("tags"); ok {
		params.Tags = strings.Join(convertToStringSlice(p.(*schema.Set).List()), ",")
	}
	if p, ok := d.GetOk("cpu_over_commit_ratio"); ok {
		params.CPUOverCommitRatio = p.(float64)
	}
	if p, ok := d.GetOk("memory_over_commit_ratio"); ok {
		params.MemoryOverCommitRatio = p.(float64)
	}
	if p, ok := d.GetOk("default_macvlan_mode"); ok {
		params.DefaultMacvlanMode = p.(string)
	}

	return &params
}
