package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/ionutbalutoiu/gomaasclient/client"
	"github.com/ionutbalutoiu/gomaasclient/entity"
)

var (
	validDnsRecordTypes = []string{"A/AAAA", "CNAME", "MX", "NS", "SRV", "SSHFP", "TXT"}
)

func resourceMaasDnsRecord() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDnsRecordCreate,
		ReadContext:   resourceDnsRecordRead,
		UpdateContext: resourceDnsRecordUpdate,
		DeleteContext: resourceDnsRecordDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, m interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected TYPE:IDENTIFIER", d.Id())
				}
				resourceType := idParts[0]
				if _, errors := validation.StringInSlice(validDnsRecordTypes, false)(resourceType, "type"); len(errors) > 0 {
					return nil, errors[0]
				}
				client := m.(*client.Client)
				resourceIdentifier := idParts[1]
				var tfState map[string]interface{}
				if resourceType == "A/AAAA" {
					dnsRecord, err := getDnsResource(client, resourceIdentifier)
					if err != nil {
						return nil, err
					}
					ips := []string{}
					for _, ipAddress := range dnsRecord.IPAddresses {
						ips = append(ips, ipAddress.IP.String())
					}
					tfState = map[string]interface{}{
						"id":   fmt.Sprintf("%v", dnsRecord.ID),
						"type": resourceType,
						"data": strings.Join(ips, " "),
						"fqdn": dnsRecord.FQDN,
						"ttl":  dnsRecord.AddressTTL,
					}
				} else {
					dnsRecord, err := getDnsResourceRecord(client, resourceIdentifier)
					if err != nil {
						return nil, err
					}
					tfState = map[string]interface{}{
						"id":   fmt.Sprintf("%v", dnsRecord.ID),
						"type": dnsRecord.RRType,
						"data": dnsRecord.RRData,
						"fqdn": dnsRecord.FQDN,
						"ttl":  dnsRecord.TTL,
					}
				}
				if err := setTerraformState(d, tfState); err != nil {
					return nil, err
				}
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(validDnsRecordTypes, false)),
			},
			"data": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"domain"},
				ExactlyOneOf: []string{"name", "fqdn"},
			},
			"domain": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"name"},
			},
			"fqdn": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "fqdn"},
			},
			"ttl": {
				Type:     schema.TypeInt,
				Optional: true,
			},
		},
	}
}

func resourceDnsRecordCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	var resourceID int
	if d.Get("type").(string) == "A/AAAA" {
		dnsRecord, err := client.DNSResources.Create(getDnsResourceParams(d))
		if err != nil {
			return diag.FromErr(err)
		}
		resourceID = dnsRecord.ID
	} else {
		dnsRecord, err := client.DNSResourceRecords.Create(getDnsResourceRecordParams(d))
		if err != nil {
			return diag.FromErr(err)
		}
		resourceID = dnsRecord.ID
	}
	d.SetId(fmt.Sprintf("%v", resourceID))

	return resourceDnsRecordUpdate(ctx, d, m)
}

func resourceDnsRecordRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if d.Get("type").(string) == "A/AAAA" {
		if _, err := client.DNSResource.Get(id); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if _, err := client.DNSResourceRecord.Get(id); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func resourceDnsRecordUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if d.Get("type").(string) == "A/AAAA" {
		if _, err := client.DNSResource.Update(id, getDnsResourceParams(d)); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if _, err := client.DNSResourceRecord.Update(id, getDnsResourceRecordParams(d)); err != nil {
			return diag.FromErr(err)
		}
	}

	return resourceDnsRecordRead(ctx, d, m)
}

func resourceDnsRecordDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*client.Client)

	id, err := strconv.Atoi(d.Id())
	if err != nil {
		return diag.FromErr(err)
	}
	if d.Get("type").(string) == "A/AAAA" {
		dnsResource, err := client.DNSResource.Get(id)
		if err != nil {
			return diag.FromErr(err)
		}
		if err := client.DNSResource.Delete(id); err != nil {
			return diag.FromErr(err)
		}
		for _, ipAddress := range dnsResource.IPAddresses {
			if err := client.IPAddresses.Release(&entity.IPAddressesParams{IP: ipAddress.IP.String()}); err != nil {
				return diag.FromErr(err)
			}
		}
	} else {
		if err := client.DNSResourceRecord.Delete(id); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func getDnsResourceParams(d *schema.ResourceData) *entity.DNSResourceParams {
	return &entity.DNSResourceParams{
		IPAddresses: d.Get("data").(string),
		Name:        d.Get("name").(string),
		Domain:      d.Get("domain").(string),
		FQDN:        d.Get("fqdn").(string),
		AddressTTL:  d.Get("ttl").(int),
	}
}

func getDnsResourceRecordParams(d *schema.ResourceData) *entity.DNSResourceRecordParams {
	return &entity.DNSResourceRecordParams{
		RRType: d.Get("type").(string),
		RRData: d.Get("data").(string),
		Name:   d.Get("name").(string),
		Domain: d.Get("domain").(string),
		FQDN:   d.Get("fqdn").(string),
		TTL:    d.Get("ttl").(int),
	}
}

func getDnsResourceRecord(client *client.Client, identifier string) (*entity.DNSResourceRecord, error) {
	dnsResourceRecords, err := client.DNSResourceRecords.Get()
	if err != nil {
		return nil, err
	}
	for _, d := range dnsResourceRecords {
		if fmt.Sprintf("%v", d.ID) == identifier || d.FQDN == identifier {
			return &d, nil
		}
	}
	return nil, fmt.Errorf("DNS resource record (%s) was not found", identifier)
}

func getDnsResource(client *client.Client, identifier string) (*entity.DNSResource, error) {
	dnsResources, err := client.DNSResources.Get()
	if err != nil {
		return nil, err
	}
	for _, d := range dnsResources {
		if fmt.Sprintf("%v", d.ID) == identifier || d.FQDN == identifier {
			return &d, nil
		}
	}
	return nil, fmt.Errorf("DNS resource (%s) was not found", identifier)
}
