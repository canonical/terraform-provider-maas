package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

func resourceMaasVMHost() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVMHostCreate,
		ReadContext:   resourceVMHostRead,
		UpdateContext: resourceVMHostUpdate,
		DeleteContext: resourceVMHostDelete,

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

func resourceVMHostCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Create VM host
	vmHost, err := client.Pods.Create(getVMHostCreateParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	// Save Id
	d.SetId(fmt.Sprintf("%v", vmHost.ID))

	// Return updated VM host
	return resourceVMHostUpdate(ctx, d, m)
}

func resourceVMHostRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Get VM host details
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	vmHost, err := client.Pod.Get(id)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set Terraform state
	if err := d.Set("name", vmHost.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("zone", vmHost.Zone.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pool", vmHost.Pool.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", vmHost.Tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_cores_available", vmHost.Available.Cores); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_cores_total", vmHost.Total.Cores); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_memory_available", vmHost.Available.Memory); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_memory_total", vmHost.Total.Memory); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_local_storage_available", vmHost.Available.LocalStorage); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("resources_local_storage_total", vmHost.Total.LocalStorage); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceVMHostUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Get the VM host
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	vmHost, err := client.Pod.Get(id)
	if err != nil {
		return diag.FromErr(err)
	}

	// Update VM host options
	_, err = client.Pod.Update(vmHost.ID, getVMHostUpdateParams(d, vmHost))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourceVMHostRead(ctx, d, m)
}

func resourceVMHostDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	// Delete VM host
	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	err = client.Pod.Delete(id)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getVMHostCreateParams(d *schema.ResourceData) *entity.PodParams {
	params := entity.PodParams{
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

func getVMHostUpdateParams(d *schema.ResourceData, vmHost *entity.Pod) *entity.PodParams {
	params := entity.PodParams{
		Type:                  vmHost.Type,
		Name:                  vmHost.Name,
		PowerAddress:          d.Get("power_address").(string),
		CPUOverCommitRatio:    vmHost.CPUOverCommitRatio,
		MemoryOverCommitRatio: vmHost.MemoryOverCommitRatio,
		DefaultMacvlanMode:    vmHost.DefaultMACVLANMode,
		Zone:                  vmHost.Zone.Name,
		Pool:                  vmHost.Pool.Name,
		Tags:                  strings.Join(vmHost.Tags, ","),
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
