package maas_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"

	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASDNSRecord_basic(t *testing.T) {
	var dnsRecord entity.DNSResource
	recordBaseName := acctest.RandomWithPrefix("tf-dns-record-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccMAASDNSRecordCheckDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: getDNSRecordConfig(recordBaseName),
				Check: resource.ComposeTestCheckFunc(
					testAccMAASDNSRecordCheckExists("maas_dns_record.test_aaaa_1", &dnsRecord),
					resource.TestCheckResourceAttr("maas_dns_record.test_aaaa_1", "name", recordBaseName+"-1"),
				),
			},
		},
	})
}

func getDNSRecordConfig(recordBaseName string) string {
	return fmt.Sprintf(`
	resource "maas_dns_record" "test_aaaa_1" {
		name = "%s-1"
		type = "A/AAAA"
		data = "8.8.8.8"
		domain = "maas"
	}
	resource "maas_dns_record" "test_aaaa_2" {
		name = "%s-2"
		type = "A/AAAA"
		data = "8.8.8.8"
		domain = "maas"
	}
	`, recordBaseName, recordBaseName)
}

// Check if the DNS record specified actually exists in MAAS
func testAccMAASDNSRecordCheckExists(rn string, dnsRecord *entity.DNSResource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Check if it exists in state
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", rn, s.RootModule().Resources)
		}	

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}	

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}	
		gotDNSRecord, err := conn.DNSResource.Get(id)
		if err != nil {
			return fmt.Errorf("error getting dns record: %s", err)
		}	

		*dnsRecord = *gotDNSRecord

		return nil
	}	
}	

func testAccMAASDNSRecordCheckDestroy(s *terraform.State) error {
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client
	for _, rs := range s.RootModule().Resources {
		// Skip if the resource is not a dns record
		if rs.Type != "maas_dns_record" {
			continue
		}
		// Convert the resource ID to an integer
		id, err := strconv.Atoi(rs.Primary.ID)
		if err != nil {
			return err
		}

		// Check if the dns record exists
		response, err := conn.DNSResource.Get(id)
		if err == nil {
			if response != nil && response.ID == id {
				return fmt.Errorf("dns record still exists: %s", rs.Primary.ID)
			}
		}

		// If the error is equivalent to 404 not found, the dns record is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "404 Not Found") {	
			return err
		}

	}
	return nil
}

