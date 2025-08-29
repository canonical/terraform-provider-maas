package maas_test

import (
	"fmt"
	"os"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"terraform-provider-maas/maas"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestNormalizeConfigValue(t *testing.T) {
	testCases := []struct {
		apiValue []byte
		expected string
		label    string
	}{
		{
			apiValue: []byte{0x6e, 0x75, 0x6c, 0x6c},
			expected: "",
			label:    "literal null",
		},
		{
			apiValue: []byte("\"null\""),
			expected: "null",
			label:    "string literal null",
		},
		{
			apiValue: []byte("hello world"),
			expected: "hello world",
			label:    "unquoted string",
		},
		{
			apiValue: []byte("example.com:514"),
			expected: "example.com:514",
			label:    "unquoted string 2",
		},
		{
			apiValue: []byte("\"example.com:514\""),
			expected: "example.com:514",
			label:    "quoted string",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.label, func(t *testing.T) {
			res := maas.NormalizeConfigValue(testCase.apiValue)
			if res != testCase.expected {
				t.Errorf("expected %q, got %q for case %q", testCase.expected, res, testCase.label)
			}
		})
	}
}

func TestAccResourceMAASConfiguration_basic(t *testing.T) {
	availableDistroSeries := os.Getenv("TF_ACC_CONFIGURATION_DISTRO_SERIES")

	// Test all MAAS Settings cases. Set twice to ensure a change is actually made
	// in case defaults change, and value 2 is selected to be a typical value to
	// ensure dev environments aren't greatly changed between test runs.
	testCases := []struct {
		key    string
		value1 string
		value2 string
	}{
		{
			key:    "active_discovery_interval",
			value1: "21600",
			value2: "10800",
		},
		{
			key:    "auto_vlan_creation",
			value1: "false",
			value2: "true",
		},
		{
			key:    "boot_images_auto_import",
			value1: "false",
			value2: "true",
		},
		{
			key:    "boot_images_no_proxy",
			value1: "true",
			value2: "false",
		},
		{
			key:    "commissioning_distro_series",
			value1: availableDistroSeries,
			value2: availableDistroSeries, // Dev environment does not necessarily have multiple distros
		},
		{
			key:    "completed_intro",
			value1: "false",
			value2: "true",
		},
		{
			key:    "curtin_verbose",
			value1: "false",
			value2: "true",
		},
		{
			key:    "default_distro_series",
			value1: availableDistroSeries,
			value2: availableDistroSeries, // Dev environment does not necessarily have multiple distros
		},
		{
			key:    "default_dns_ttl",
			value1: "60",
			value2: "30",
		},
		{
			key:    "default_osystem",
			value1: "",
			value2: "ubuntu",
		},
		{
			key:    "default_storage_layout",
			value1: "bcache",
			value2: "blank",
		},
		{
			key:    "disk_erase_with_quick_erase",
			value1: "true",
			value2: "false",
		},
		{
			key:    "disk_erase_with_secure_erase",
			value1: "true",
			value2: "false",
		},
		{
			key:    "dns_trusted_acl",
			value1: "192.168.1.1 192.168.1.2",
			value2: "",
		},
		{
			key:    "dnssec_validation",
			value1: "yes",
			value2: "auto",
		},
		{
			key:    "enable_analytics",
			value1: "false",
			value2: "true",
		},
		{
			key:    "enable_disk_erasing_on_release",
			value1: "true",
			value2: "false",
		},
		{
			key:    "enable_http_proxy",
			value1: "false",
			value2: "true",
		},
		{
			key:    "enable_third_party_drivers",
			value1: "false",
			value2: "true",
		},
		{
			key:    "enlist_commissioning",
			value1: "false",
			value2: "true",
		},
		{
			key:    "force_v1_network_yaml",
			value1: "true",
			value2: "false",
		},
		{
			key:    "hardware_sync_interval",
			value1: "10m",
			value2: "15m",
		},
		{
			key:    "http_proxy",
			value1: "http://proxy.example.com:8080",
			value2: "",
		},
		{
			key:    "kernel_opts",
			value1: "console=ttyS0",
			value2: "",
		},
		{
			key:    "maas_auto_ipmi_cipher_suite_id",
			value1: "8",
			value2: "3",
		},
		{
			key:    "maas_auto_ipmi_k_g_bmc_key",
			value1: "12345678901234567890", // Must be 20 characters
			value2: "",
		},
		{
			key:    "maas_auto_ipmi_user",
			value1: "admin",
			value2: "maas",
		},
		{
			key:    "maas_auto_ipmi_user_privilege_level",
			value1: "USER",
			value2: "ADMIN",
		},
		// {
		// 	key: "maas_auto_ipmi_workaround_flags",   // Will work when this bug is fixed https://bugs.launchpad.net/maas/+bug/2112191
		// 	value1:  "asdf",
		// },
		{
			key:    "maas_internal_domain",
			value1: "maas-internal-alt",
			value2: "maas-internal",
		},
		{
			key:    "maas_name",
			value1: "maas-alt",
			value2: "maas-dev",
		},
		{
			key:    "maas_proxy_port",
			value1: "8080",
			value2: "8000",
		},
		{
			key:    "maas_syslog_port",
			value1: "49152",
			value2: "5247",
		},
		{
			key:    "max_node_commissioning_results",
			value1: "100",
			value2: "10",
		},
		{
			key:    "max_node_installation_results",
			value1: "5",
			value2: "3",
		},
		{
			key:    "max_node_release_results",
			value1: "5",
			value2: "3",
		},
		{
			key:    "max_node_testing_results",
			value1: "100",
			value2: "10",
		},
		{
			key:    "network_discovery",
			value1: "disabled",
			value2: "enabled",
		},
		{
			key:    "node_timeout",
			value1: "10",
			value2: "30",
		},
		{
			key:    "ntp_external_only",
			value1: "true",
			value2: "false",
		},
		{
			key:    "ntp_servers",
			value1: "ntp.example.com",
			value2: "ntp.ubuntu.com",
		},
		{
			key:    "prefer_v4_proxy",
			value1: "true",
			value2: "false",
		},
		{
			key:    "prometheus_enabled",
			value1: "true",
			value2: "false",
		},
		{
			key:    "prometheus_push_gateway",
			value1: "http://prometheus.example.com:9090",
			value2: "",
		},
		{
			key:    "prometheus_push_interval",
			value1: "80",
			value2: "60",
		},
		{
			key:    "promtail_enabled",
			value1: "true",
			value2: "false",
		},
		{
			key:    "promtail_port",
			value1: "49153",
			value2: "5238",
		},
		{
			key:    "release_notifications",
			value1: "false",
			value2: "true",
		},
		{
			key:    "remote_syslog",
			value1: "example.com:514",
			value2: "", // Returning null if unset
		},
		{
			key:    "session_length",
			value1: "604800",  // 7 days
			value2: "1209600", // 14 days
		},
		{
			key:    "subnet_ip_exhaustion_threshold_count",
			value1: "24",
			value2: "16",
		},
		{
			key:    "theme",
			value1: "sage",
			value2: "",
		},
		{
			key:    "tls_cert_expiration_notification_enabled",
			value1: "true",
			value2: "false",
		},
		{
			key:    "tls_cert_expiration_notification_interval",
			value1: "60",
			value2: "30",
		},
		{
			key:    "upstream_dns",
			value1: "8.8.8.8",
			value2: "",
		},
		{
			key:    "use_peer_proxy",
			value1: "true",
			value2: "false",
		},
		{
			key:    "use_rack_proxy",
			value1: "false",
			value2: "true",
		},
		{
			key:    "vcenter_datacenter",
			value1: "maas-vcenter",
			value2: "",
		},
		{
			key:    "vcenter_password",
			value1: "dummy-password",
			value2: "",
		},
		{
			key:    "vcenter_server",
			value1: "vcenter.example.com",
			value2: "",
		},
		{
			key:    "vcenter_username",
			value1: "maas",
			value2: "",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.key, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				PreCheck:     func() { testutils.PreCheck(t, []string{"TF_ACC_CONFIGURATION_DISTRO_SERIES"}) },
				Providers:    testutils.TestAccProviders,
				ErrorCheck:   func(err error) error { return err },
				CheckDestroy: testAccMAASConfigurationCheckDestroy,
				Steps: []resource.TestStep{
					{
						Config: testAccMAASConfigurationConfigBasic(testCase.key, testCase.value1),
						Check: resource.ComposeTestCheckFunc(
							testAccMAASConfigurationCheckExists("maas_configuration.test"),
							resource.TestCheckResourceAttr("maas_configuration.test", "key", testCase.key),
							resource.TestCheckResourceAttr("maas_configuration.test", "value", testCase.value1),
						),
					},
					{
						Config: testAccMAASConfigurationConfigBasic(testCase.key, testCase.value2),
						Check: resource.ComposeTestCheckFunc(
							testAccMAASConfigurationCheckExists("maas_configuration.test"),
							resource.TestCheckResourceAttr("maas_configuration.test", "key", testCase.key),
							resource.TestCheckResourceAttr("maas_configuration.test", "value", testCase.value2),
						),
					},
				},
			})
		})
	}
}

func testAccMAASConfigurationCheckExists(rn string) resource.TestCheckFunc {
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

		value, err := conn.MAASServer.Get(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error getting configuration value: %s", err)
		}

		cleanedVal := maas.NormalizeConfigValue(value)
		if cleanedVal != rs.Primary.Attributes["value"] {
			return fmt.Errorf("configuration value does not match: expected %s, got %s", rs.Primary.Attributes["value"], value)
		}

		return nil
	}
}

// Not required for this resource as no resources are created, but is required by the linter.
func testAccMAASConfigurationCheckDestroy(s *terraform.State) error {
	return nil
}

func testAccMAASConfigurationConfigBasic(key string, value string) string {
	return fmt.Sprintf(`
resource "maas_configuration" "test" {
  key   = %q
  value = %q
}
`, key, value)
}
