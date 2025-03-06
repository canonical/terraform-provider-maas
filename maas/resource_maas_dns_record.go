package maas

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/go-set/v2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	validDnsRecordTypes = []string{"A/AAAA", "CNAME", "MX", "NS", "SRV", "SSHFP", "TXT"}
)

func resourceMaasDnsRecord() *schema.Resource {
	return &schema.Resource{
		Description:   "Provides a resource to manage MAAS DNS domain records.",
		CreateContext: resourceDnsRecordCreate,
		ReadContext:   resourceDnsRecordRead,
		UpdateContext: resourceDnsRecordUpdate,
		DeleteContext: resourceDnsRecordDelete,
		Importer: &schema.ResourceImporter{
			StateContext: func(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), ":")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("unexpected format of ID (%q), expected TYPE:IDENTIFIER", d.Id())
				}
				resourceType := idParts[0]
				if _, errors := validation.StringInSlice(validDnsRecordTypes, false)(resourceType, "type"); len(errors) > 0 {
					return nil, errors[0]
				}
				client := meta.(*ClientConfig).Client

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
			"data": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The data set for the new DNS record.",
			},
			"domain": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"name"},
				Description:  "The domain of the new DNS record. Used in conjunction with `name`. It conflicts with `fqdn` argument.",
			},
			"fqdn": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "fqdn"},
				Description:  "The fully qualified domain name of the new DNS record. This contains the name and the domain of the new DNS record. It conflicts with `name` and `domain` arguments.",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"domain"},
				ExactlyOneOf: []string{"name", "fqdn"},
				Description:  "The new DNS record resource name. Used in conjunction with `domain`. It conflicts with `fqdn` argument.",
			},
			"ttl": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "The TTL of the new DNS record.",
			},
			"type": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice(validDnsRecordTypes, false)),
				Description:      "The DNS record type. Valid options are: `A/AAAA`, `CNAME`, `MX`, `NS`, `SRV`, `SSHFP`, `TXT`.",
			},
		},
	}
}

func resourceDnsRecordCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

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

	return resourceDnsRecordUpdate(ctx, d, meta)
}

func resourceDnsRecordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

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

func resourceDnsRecordUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

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

	return resourceDnsRecordRead(ctx, d, meta)
}

func resourceDnsRecordDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ClientConfig).Client

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
		if err := releaseDNSResourceIPAddresses(client, dnsResource, id); err != nil {
			return diag.FromErr(err)
		}
	} else {
		if err := client.DNSResourceRecord.Delete(id); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

// Release all IP addresses of a DNS resource, but only if it is not used by other DNS resources.
func releaseDNSResourceIPAddresses(client *client.Client, dnsResource *entity.DNSResource, dnsID int) error {
	allDNSResources, err := client.DNSResources.Get(&entity.DNSResourcesParams{})
	if err != nil {
		return err
	}
	allOtherDNSResourcesIPAddresses := set.New[string](0)
	for _, r := range allDNSResources {
		if r.ID == dnsID {
			continue
		}
		for _, ipAddress := range r.IPAddresses {
			allOtherDNSResourcesIPAddresses.Insert(ipAddress.IP.String())
		}
	}
	for _, ipAddress := range dnsResource.IPAddresses {
		if allOtherDNSResourcesIPAddresses.Contains(ipAddress.IP.String()) {
			continue
		}
		// Release the IP address if it is not used by another DNS resource.
		// Terraform removes resources in parallel, so if it's already been deleted
		// by another resource pointing to the same IP being removed, ignore it.
		if err := client.IPAddresses.Release(&entity.IPAddressesParams{IP: ipAddress.IP.String()}); err != nil && !strings.Contains(err.Error(), "does not exist") {
			return err
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
	dnsResourceRecords, err := client.DNSResourceRecords.Get(&entity.DNSResourceRecordsParams{})
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
	dnsResources, err := client.DNSResources.Get(&entity.DNSResourcesParams{})
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
