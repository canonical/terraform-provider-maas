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
	recordName := acctest.RandomWithPrefix("tf-dns-record-")
	const recordName1 = "maas_dns_record.test"
	const TEST_IP_ADDRESS = "8.8.8.8"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccMAASDNSRecordCheckDestroy(TEST_IP_ADDRESS),
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: getDNSRecordConfigBasic(recordName, "A/AAAA", TEST_IP_ADDRESS, "maas"),
				Check: resource.ComposeTestCheckFunc(
					testAccMAASDNSRecordCheckExists(recordName1, &dnsRecord),
					resource.TestCheckResourceAttr(recordName1, "name", recordName),
					resource.TestCheckResourceAttr(recordName1, "type", "A/AAAA"),
					resource.TestCheckResourceAttr(recordName1, "data", TEST_IP_ADDRESS),
					resource.TestCheckResourceAttr(recordName1, "domain", "maas"),
				),
			},
		},
	})
}

// Test that two DNS records with the same IP address can be created and destroyed.
func TestAccResourceMAASDNSRecord_same_ip_address(t *testing.T) {
	var dnsRecord entity.DNSResource
	recordBaseName := acctest.RandomWithPrefix("tf-dns-record-")
	const recordName1 = "maas_dns_record.test_aaaa_1"
	const recordName2 = "maas_dns_record.test_aaaa_2"
	const TEST_IP_ADDRESS_2 = "8.8.8.9"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccMAASDNSRecordCheckDestroy(TEST_IP_ADDRESS_2),
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: getDNSRecordConfigSameIPA_AAAA(recordBaseName, TEST_IP_ADDRESS_2),
				Check: resource.ComposeTestCheckFunc(
					testAccMAASDNSRecordCheckExists(recordName1, &dnsRecord),
					testAccMAASDNSRecordCheckExists(recordName2, &dnsRecord),
					// resource.TestCheckResourceAttr(recordName1, "name", recordBaseName+"-1"),
					// resource.TestCheckResourceAttr(recordName2, "name", recordBaseName+"-2"),
					resource.TestCheckResourceAttr(recordName1, "type", "A/AAAA"),
					resource.TestCheckResourceAttr(recordName2, "type", "A/AAAA"),
					resource.TestCheckResourceAttr(recordName1, "data", TEST_IP_ADDRESS_2),
					resource.TestCheckResourceAttr(recordName2, "data", TEST_IP_ADDRESS_2),
					resource.TestCheckResourceAttr(recordName1, "domain", "maas"),
					resource.TestCheckResourceAttr(recordName2, "domain", "maas"),
				),
			},
		},
	})
}

func getDNSRecordConfigSameIPA_AAAA(recordBaseName string, ipAddress string) string {
	return fmt.Sprintf(`
	resource "maas_dns_record" "test_aaaa_1" {
	  name   = "%s-1"
	  type   = "A/AAAA"
	  data   = %q
	  domain = "maas"
	}
	resource "maas_dns_record" "test_aaaa_2" {
	  name   = "%s-2"
	  type   = "A/AAAA"
	  data   = %q
	  domain = "maas"
	}
	`, recordBaseName, ipAddress, recordBaseName, ipAddress)
}

func getDNSRecordConfigBasic(name string, recordType string, data string, domain string) string {
	return fmt.Sprintf(`
	resource "maas_dns_record" "test" {
	  name   = "%s"
	  type   = "%s"
	  data   = "%s"
	  domain = "%s"
	}
	`, name, recordType, data, domain)
}

// Check if the DNS record specified exists in MAAS. If not, return an error.
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
			return fmt.Errorf("ip address is not released: %s with error: %s", ipAddress, err	)
		}
		return nil
	}
}

// Check if a particular IP address is allocated in MAAS
func isIPAddressAllocated(conn *client.Client, ipAddress string) (bool, error) {
	params := &entity.IPAddressesParams{IP: ipAddress}
	allIPAddresses, err := conn.IPAddresses.Get(params)
	if err != nil {
		if strings.Contains(err.Error(), "does not exist") {
			// The IP address is already allocated
			return false, nil
		}
		// The IP address could be allocated, return the error
		return false, err 
	}

	for _, ip := range allIPAddresses {
		if ip.IP.String() == ipAddress {
			return true, fmt.Errorf("ip address is allocated")
		}
	}
	return false, nil
}
