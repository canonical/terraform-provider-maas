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

var (
	defaultBlockSize    = 512
	defaultIsBootDevice = false
)

func resourceMaasBlockDevice() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceBlockDeviceCreate,
		ReadContext:   resourceBlockDeviceRead,
		UpdateContext: resourceBlockDeviceUpdate,
		DeleteContext: resourceBlockDeviceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected MACHINE:BLOCK_DEVICE", d.Id())
				}
				client := m.(*client.Client)
				machine, err := findMachine(client, idParts[0])
				if err != nil {
					return nil, err
				}
				blockDevice, err := findBlockDevice(client, machine.SystemID, idParts[1])
				if err != nil {
					return nil, err
				}
				if blockDevice == nil {
					return nil, fmt.Errorf("block device (%s) was not found on machine (%s)", idParts[1], machine.Hostname)
				}
				if err := d.Set("machine", machine.SystemID); err != nil {
					return nil, err
				}
				if err := d.Set("name", blockDevice.Name); err != nil {
					return nil, err
				}
				sizeGigabytes := blockDevice.Size / (1024 * 1024 * 1024)
				if err := d.Set("size_gigabytes", int(sizeGigabytes)); err != nil {
					return nil, err
				}
				if err := d.Set("block_size", defaultBlockSize); err != nil {
					return nil, err
				}
				if err := d.Set("is_boot_device", defaultIsBootDevice); err != nil {
					return nil, err
				}
				d.SetId(fmt.Sprintf("%v", blockDevice.ID))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"machine": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"size_gigabytes": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"block_size": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  defaultBlockSize,
			},
			"is_boot_device": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  defaultIsBootDevice,
			},
			"partitions": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"size_gigabytes": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"bootable": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"tags": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"fs_type": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"label": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mount_point": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mount_options": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"path": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"model": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				RequiredWith:  []string{"serial"},
				ConflictsWith: []string{"id_path"},
				AtLeastOneOf:  []string{"model", "id_path"},
			},
			"serial": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				RequiredWith:  []string{"model"},
				ConflictsWith: []string{"id_path"},
			},
			"id_path": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"model", "serial"},
				AtLeastOneOf:  []string{"model", "id_path"},
			},
			"tags": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"uuid": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"path": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceBlockDeviceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	machine, err := findMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	blockDevice, err := findBlockDevice(client, machine.SystemID, d.Get("name").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if blockDevice == nil {
		blockDevice, err = client.BlockDevices.Create(machine.SystemID, getBlockDeviceParams(d))
		if err != nil {
			return diag.FromErr(err)
		}
	}
	d.SetId(fmt.Sprintf("%v", blockDevice.ID))

	return resourceBlockDeviceUpdate(ctx, d, m)
}

func resourceBlockDeviceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	machine, err := findMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	blockDevice, err := client.BlockDevice.Get(machine.SystemID, id)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("model", blockDevice.Model); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("serial", blockDevice.Serial); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("id_path", blockDevice.IDPath); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("tags", blockDevice.Tags); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("uuid", blockDevice.UUID); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("path", blockDevice.IDPath); err != nil {
		return diag.FromErr(err)
	}
	if err := setBlockDevicePartitionsTFState(d, blockDevice); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceBlockDeviceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	machine, err := findMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	blockDevice, err := client.BlockDevice.Update(machine.SystemID, id, getBlockDeviceParams(d))
	if err != nil {
		return diag.FromErr(err)
	}

	if err := setBlockDeviceTags(client, d, blockDevice); err != nil {
		return diag.FromErr(err)
	}

	if p, ok := d.GetOk("is_boot_device"); ok {
		if p.(bool) {
			if err := client.BlockDevice.SetBootDisk(machine.SystemID, id); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if err := createBlockDevicePartitions(client, d, blockDevice); err != nil {
		return diag.FromErr(err)
	}

	return resourceBlockDeviceRead(ctx, d, m)
}

func resourceBlockDeviceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	machine, err := findMachine(client, d.Get("machine").(string))
	if err != nil {
		return diag.FromErr(err)
	}
	if err := client.BlockDevice.Delete(machine.SystemID, id); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func getBlockDeviceParams(d *schema.ResourceData) *entity.BlockDeviceParams {
	params := entity.BlockDeviceParams{
		Name:      d.Get("name").(string),
		Size:      d.Get("size_gigabytes").(int) * 1024 * 1024 * 1024,
		BlockSize: d.Get("block_size").(int),
	}
	if p, ok := d.GetOk("model"); ok {
		params.Model = p.(string)
	}
	if p, ok := d.GetOk("serial"); ok {
		params.Serial = p.(string)
	}
	if p, ok := d.GetOk("id_path"); ok {
		params.IDPath = p.(string)
	}
	return &params
}

func findBlockDevice(client *client.Client, machineID string, identifier string) (*entity.BlockDevice, error) {
	blockDevices, err := client.BlockDevices.Get(machineID)
	if err != nil {
		return nil, err
	}

	for _, b := range blockDevices {
		if fmt.Sprintf("%v", b.ID) == identifier || b.Name == identifier || b.IDPath == identifier || b.Path == identifier {
			return &b, nil
		}
	}

	return nil, nil
}

func setBlockDeviceTags(client *client.Client, d *schema.ResourceData, blockDevice *entity.BlockDevice) error {
	p, ok := d.GetOk("tags")
	if !ok {
		return nil
	}
	tags := p.(*schema.Set).List()
	blockDevice, err := client.BlockDevice.Get(blockDevice.SystemID, blockDevice.ID)
	if err != nil {
		return err
	}
	// Remove existing tags
	for _, t := range blockDevice.Tags {
		if _, err = client.BlockDevice.RemoveTag(blockDevice.SystemID, blockDevice.ID, t); err != nil {
			return err
		}
	}
	// Add new tags
	for _, t := range tags {
		if _, err = client.BlockDevice.AddTag(blockDevice.SystemID, blockDevice.ID, t.(string)); err != nil {
			return err
		}
	}
	return nil
}

func setBlockDevicePartitionsTFState(d *schema.ResourceData, blockDevice *entity.BlockDevice) error {
	partitions := make([]map[string]interface{}, len(blockDevice.Partitions))
	for i, p := range blockDevice.Partitions {
		part := map[string]interface{}{
			"size_gigabytes": int(p.Size / (1024 * 1024 * 1024)),
			"bootable":       p.Bootable,
			"tags":           p.Tags,
			"path":           p.Path,
			"fs_type":        p.FileSystem.FSType,
			"label":          p.FileSystem.Label,
			"mount_point":    p.FileSystem.MountPoint,
			"mount_options":  p.FileSystem.MountOptions,
		}
		partitions[i] = part
	}
	if err := d.Set("partitions", partitions); err != nil {
		return err
	}
	return nil
}

func createBlockDevicePartitions(client *client.Client, d *schema.ResourceData, blockDevice *entity.BlockDevice) error {
	p, ok := d.GetOk("partitions")
	if !ok {
		return nil
	}
	// Remove existing partitions
	for _, part := range blockDevice.Partitions {
		if err := client.BlockDevicePartition.Delete(blockDevice.SystemID, blockDevice.ID, part.ID); err != nil {
			return err
		}

	}
	// Create new partitions given by the user
	partitions := p.([]interface{})
	for _, part := range partitions {
		partition := part.(map[string]interface{})
		partitionParams := entity.BlockDevicePartitionParams{
			Size:     partition["size_gigabytes"].(int) * 1024 * 1024 * 1024,
			Bootable: partition["bootable"].(bool),
		}
		blockDevicePartition, err := client.BlockDevicePartitions.Create(blockDevice.SystemID, blockDevice.ID, &partitionParams)
		if err != nil {
			return err
		}
		tags := partition["tags"].(*schema.Set).List()
		for _, t := range tags {
			if _, err := client.BlockDevicePartition.AddTag(blockDevice.SystemID, blockDevice.ID, blockDevicePartition.ID, t.(string)); err != nil {
				return err
			}
		}
		if fsType := partition["fs_type"].(string); fsType != "" {
			label := partition["label"].(string)
			if _, err := client.BlockDevicePartition.Format(blockDevice.SystemID, blockDevice.ID, blockDevicePartition.ID, fsType, label); err != nil {
				return err
			}
			if mountPoint := partition["mount_point"].(string); mountPoint != "" {
				mountOptions := partition["mount_options"].(string)
				if _, err := client.BlockDevicePartition.Mount(blockDevice.SystemID, blockDevice.ID, blockDevicePartition.ID, mountPoint, mountOptions); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
