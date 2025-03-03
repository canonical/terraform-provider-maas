package maas_test

import (
	"fmt"
	"strconv"
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
	randomString := acctest.RandomWithPrefix("tf-dns-record-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: func(s *terraform.State) error { return nil },
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: getDNSRecordConfig(randomString),
				Check: resource.ComposeTestCheckFunc(
					testAccMAASDNSRecordCheckExists("maas_dns_record.test_aaaa_1", &dnsRecord),
				),
			},
		},
	})
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

func getDNSRecordConfig(randomString string) string {
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
	`, randomString, randomString)
}

