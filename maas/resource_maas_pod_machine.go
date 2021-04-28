package maas

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ionutbalutoiu/gomaasclient/api/endpoint"
	"github.com/ionutbalutoiu/gomaasclient/gmaw"
	"github.com/ionutbalutoiu/gomaasclient/maas"
	"github.com/juju/gomaasapi"
)

func resourceMaasPodMachine() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePodMachineCreate,
		ReadContext:   resourcePodMachineRead,
		UpdateContext: resourcePodMachineUpdate,
		DeleteContext: resourcePodMachineDelete,

		Schema: map[string]*schema.Schema{
			"pod": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cores": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"pinned_cores": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"storage": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"interfaces": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"domain": {
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
		},
	}
}

func resourcePodMachineCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*gomaasapi.MAASObject)

	// Find Pod
	pod, err := findPod(client, d.Get("pod").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	podManager, err := maas.NewPodManager(pod.ID, gmaw.NewPod(client))
	if err != nil {
		return diag.FromErr(err)
	}

	// Create Pod machine
	params := getPodMachineCreateParams(d)
	machine, err := podManager.Compose(params)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set Terraform state
	if err := d.Set("cores", params.Cores); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pinned_cores", params.PinnedCores); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("memory", params.Memory); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("storage", params.Storage); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("interfaces", params.Interfaces); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(machine.SystemID)

	// Wait for Pod machine to be ready
	log.Printf("[DEBUG] Waiting for machine (%s) to become ready\n", machine.SystemID)
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"Commissioning", "Testing"},
		Target:     []string{"Ready"},
		Refresh:    getMachineStatusFunc(client, machine.SystemID),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}
	_, err = stateConf.WaitForStateContext(ctx)
	if err != nil {
		return diag.FromErr(fmt.Errorf("machine (%s) didn't become ready within allowed timeout: %s", machine.SystemID, err))
	}

	// Return updated Pod machine
	return resourcePodMachineUpdate(ctx, d, m)
}

func resourcePodMachineRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*gomaasapi.MAASObject)

	// Get Pod machine
	machineManager, err := maas.NewMachineManager(d.Id(), gmaw.NewMachine(client))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get machine (%s) manager: %s", d.Id(), err))
	}
	machine := machineManager.Current()

	// Set Terraform state
	if err := d.Set("hostname", machine.Hostname); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("domain", machine.Domain.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("zone", machine.Zone.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("pool", machine.Pool.Name); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourcePodMachineUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*gomaasapi.MAASObject)

	// Update Pod machine
	machineManager, err := maas.NewMachineManager(d.Id(), gmaw.NewMachine(client))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get machine (%s) manager: %s", d.Id(), err))
	}
	err = machineManager.Update(getPodMachineUpdateParams(d, machineManager.Current()))
	if err != nil {
		return diag.FromErr(err)
	}

	return resourcePodMachineRead(ctx, d, m)
}

func resourcePodMachineDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*gomaasapi.MAASObject)

	// Delete Pod machine
	machineManager, err := maas.NewMachineManager(d.Id(), gmaw.NewMachine(client))
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to get machine (%s) manager: %s", d.Id(), err))
	}
	err = machineManager.Delete()
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func findPod(client *gomaasapi.MAASObject, podIdentifier string) (*endpoint.Pod, error) {
	podsManager := maas.NewPodsManager(gmaw.NewPods(client))
	pods, err := podsManager.Get()
	if err != nil {
		return nil, err
	}

	for _, pod := range pods {
		if fmt.Sprintf("%v", pod.ID) == podIdentifier || pod.Name == podIdentifier {
			return pod, err
		}
	}

	return nil, fmt.Errorf("pod (%s) not found", podIdentifier)
}

func getPodMachineCreateParams(d *schema.ResourceData) *endpoint.PodMachineParams {
	params := endpoint.PodMachineParams{}

	if p, ok := d.GetOk("cores"); ok {
		params.Cores = p.(int)
	}
	if p, ok := d.GetOk("pinned_cores"); ok {
		params.PinnedCores = p.(int)
	}
	if p, ok := d.GetOk("memory"); ok {
		params.Memory = p.(int)
	}
	if p, ok := d.GetOk("storage"); ok {
		params.Storage = p.(string)
	}
	if p, ok := d.GetOk("interfaces"); ok {
		params.Interfaces = p.(string)
	}
	if p, ok := d.GetOk("hostname"); ok {
		params.Hostname = p.(string)
	}

	return &params
}

func getPodMachineUpdateParams(d *schema.ResourceData, machine *endpoint.Machine) *endpoint.MachineParams {
	params := endpoint.MachineParams{
		CPUCount:     machine.CPUCount,
		Memory:       machine.Memory,
		SwapSize:     machine.SwapSize,
		Architecture: machine.Architecture,
		MinHWEKernel: machine.MinHWEKernel,
		PowerType:    machine.PowerType,
		Description:  machine.Description,
	}

	if p, ok := d.GetOk("hostname"); ok {
		params.Hostname = p.(string)
	}
	if p, ok := d.GetOk("domain"); ok {
		params.Domain = p.(string)
	}
	if p, ok := d.GetOk("zone"); ok {
		params.Zone = p.(string)
	}
	if p, ok := d.GetOk("pool"); ok {
		params.Pool = p.(string)
	}

	return &params
}
