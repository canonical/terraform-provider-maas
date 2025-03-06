package maas_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccResourceMAASDNSRecord_basic(t *testing.T) {
	var dnsRecord entity.DNSResource
	recordName := acctest.RandomWithPrefix("tf")
	resourceName := "test"
	testIPAddress := "8.8.8.8"
	testDomain := acctest.RandomWithPrefix("tf")
	testRecordType := "A/AAAA"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccMAASDNSRecordCheckDestroy(testIPAddress),
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: getDNSRecordConfigBasic(recordName, testRecordType, testIPAddress, testDomain),
				Check: resource.ComposeTestCheckFunc(
					testAccMAASDNSRecordCheckExists("maas_dns_record."+resourceName, &dnsRecord),
					resource.TestCheckResourceAttr("maas_dns_record."+resourceName, "name", recordName),
					resource.TestCheckResourceAttr("maas_dns_record."+resourceName, "type", testRecordType),
					resource.TestCheckResourceAttr("maas_dns_record."+resourceName, "data", testIPAddress),
					resource.TestCheckResourceAttr("maas_dns_record."+resourceName, "domain", testDomain),
				),
			},
		},
	})
}

// Test that two DNS records with the same IP address can be created and destroyed.
func TestAccResourceMAASDNSRecord_sameIPAddress(t *testing.T) {
	var dnsRecord entity.DNSResource
	recordName1 := acctest.RandomWithPrefix("tf-1")
	recordName2 := acctest.RandomWithPrefix("tf-2")
	resourceName1 := "test_aaaa_1"
	resourceName2 := "test_aaaa_2"
	testIPAddress := "8.8.8.9"
	testDomain := acctest.RandomWithPrefix("tf")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccMAASDNSRecordCheckDestroy(testIPAddress),
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: getDNSRecordConfigSameIPA_AAAA(testDomain, resourceName1, resourceName2, recordName1, recordName2, testIPAddress),
				Check: resource.ComposeTestCheckFunc(
					testAccMAASDNSRecordCheckExists("maas_dns_record."+resourceName1, &dnsRecord),
					testAccMAASDNSRecordCheckExists("maas_dns_record."+resourceName2, &dnsRecord),
					resource.TestCheckResourceAttr("maas_dns_record."+resourceName1, "name", recordName1),
					resource.TestCheckResourceAttr("maas_dns_record."+resourceName2, "name", recordName2),
					resource.TestCheckResourceAttr("maas_dns_record."+resourceName1, "type", "A/AAAA"),
					resource.TestCheckResourceAttr("maas_dns_record."+resourceName2, "type", "A/AAAA"),
					resource.TestCheckResourceAttr("maas_dns_record."+resourceName1, "data", testIPAddress),
					resource.TestCheckResourceAttr("maas_dns_record."+resourceName2, "data", testIPAddress),
					resource.TestCheckResourceAttr("maas_dns_record."+resourceName1, "domain", testDomain),
					resource.TestCheckResourceAttr("maas_dns_record."+resourceName2, "domain", testDomain),
				),
			},
		},
	})
}

func getDNSRecordConfigSameIPA_AAAA(domain string, resourceName1 string, resourceName2 string, recordName1 string, recordName2 string, ipAddress string) string {
	return fmt.Sprintf(`
	resource "maas_dns_domain" "test" {
	  name = %q
	}
	resource "maas_dns_record" %q {
	  name   = %q
	  type   = "A/AAAA"
	  data   = %q
	  domain = maas_dns_domain.test.name
	}
	resource "maas_dns_record" %q {
	  name   = %q
	  type   = "A/AAAA"
	  data   = %q
	  domain = maas_dns_domain.test.name
	}
	`, domain, resourceName1, recordName1, ipAddress, resourceName2, recordName2, ipAddress)
}

func getDNSRecordConfigBasic(name string, recordType string, data string, domain string) string {
	return fmt.Sprintf(`
	resource "maas_dns_domain" "test" {
	  name = %q
	}
	resource "maas_dns_record" "test" {
	  name   = "%s"
	  type   = "%s"
	  data   = "%s"
	  domain = maas_dns_domain.test.name
	}
	`, domain, name, recordType, data)
}

// Check if the DNS record specified exists in MAAS. If not, return an error.
func testAccMAASDNSRecordCheckExists(rn string, dnsRecord *entity.DNSResource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Check if it exists in state
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %v", rn, s.RootModule().Resources)
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

func testAccMAASDNSRecordCheckDestroy(ipAddress string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
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

		// Check if the IP address is released
		ipAllocated, err := isIPAddressAllocated(conn, ipAddress)
		if err != nil {
			return fmt.Errorf("error checking if ip address is released: %s", err)
		}
		if ipAllocated {
			return fmt.Errorf("ip address is not released: %s", ipAddress)
		}
		return nil
	}
}

// Check if a particular IP address is allocated in MAAS
func isIPAddressAllocated(conn *client.Client, ipAddress string) (bool, error) {
	params := &entity.IPAddressesParams{IP: ipAddress}
	maasIPAddress, err := conn.IPAddresses.Get(params)
	if err != nil {
		// Unexpected error
		return false, err
	}
	// The IP address is not allocated.
	if len(maasIPAddress) == 0 {
		return false, nil
	}
	// More than one IP address found unexpectedly
	if len(maasIPAddress) > 1 {
		return false, fmt.Errorf("more than one IP address found for %s", ipAddress)
	}
	// The IP address is allocated
	if len(maasIPAddress) == 1 && maasIPAddress[0].IP.String() == ipAddress {
		return true, nil
	}
	// Unexpected error
	return false, fmt.Errorf("unexpected error, IP address got from client is not the expected one: %v", maasIPAddress)
}
