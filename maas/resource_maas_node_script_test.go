package maas_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"terraform-provider-maas/maas"
	"terraform-provider-maas/maas/testutils"
	"testing"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
)

func TestAccResourceMAASNodeScript_basic(t *testing.T) {
	var nodeScript entity.NodeScript

	name := acctest.RandomWithPrefix("tf-node-script-")

	scriptType := "commissioning"
	title := "Initial Title"
	description := "initial description"
	comment := "initial comment"
	parallel := "instance"
	timeout := "00:10:00"
	hardwareType := "node"
	applyConfiguredNetworking := true
	destructive := true
	mayReboot := true
	recommission := true
	forHardware := []string{"system_vendor:canonical", "system_product:maas"}
	tags := []string{"dummy", "script"}
	packages := map[string][]string{"snap": {"maas", "maas-test-db"}}

	updatedScriptType := "testing"
	updatedTitle := "Updated Title"
	updatedDescription := "updated description"
	updatedComment := "updated comment"
	updatedParallel := "any"
	updatedTimeout := "00:20:00"
	updatedHardwareType := "storage"
	updatedApplyConfiguredNetworking := false
	updatedDestructive := false
	updatedMayReboot := false
	updatedRecommission := false
	updatedForHardware := []string{"system_vendor:canonical", "system_product:maas-testdb"}
	updatedTags := []string{"dummy", "script"}
	updatedPackages := map[string][]string{"snap": {"maas", "maas-test-db"}}

	scriptRaw := `
#!/bin/bash
echo "Hello World"
`
	encodedScript := base64.StdEncoding.EncodeToString([]byte(scriptRaw))

	checks := []resource.TestCheckFunc{
		testAccMAASNodeScriptCheckExists("maas_node_script.test", &nodeScript),
		resource.TestCheckResourceAttr("maas_node_script.test", "script", encodedScript),
		resource.TestCheckResourceAttr("maas_node_script.test", "name", name),
		resource.TestCheckResourceAttr("maas_node_script.test", "script_type", scriptType),
		resource.TestCheckResourceAttr("maas_node_script.test", "title", title),
		resource.TestCheckResourceAttr("maas_node_script.test", "description", description),
		resource.TestCheckResourceAttr("maas_node_script.test", "comment", comment),
		resource.TestCheckResourceAttr("maas_node_script.test", "parallel", parallel),
		resource.TestCheckResourceAttr("maas_node_script.test", "timeout", timeout),
		resource.TestCheckResourceAttr("maas_node_script.test", "hardware_type", hardwareType),
		resource.TestCheckResourceAttr("maas_node_script.test", "apply_configured_networking", fmt.Sprintf("%t", applyConfiguredNetworking)),
		resource.TestCheckResourceAttr("maas_node_script.test", "destructive", fmt.Sprintf("%t", destructive)),
		resource.TestCheckResourceAttr("maas_node_script.test", "may_reboot", fmt.Sprintf("%t", mayReboot)),
		resource.TestCheckResourceAttr("maas_node_script.test", "recommission", fmt.Sprintf("%t", recommission)),
		resource.TestCheckResourceAttr("maas_node_script.test", "tags.#", fmt.Sprintf("%v", len(tags))),
		resource.TestCheckResourceAttr("maas_node_script.test", "for_hardware.#", fmt.Sprintf("%v", len(forHardware))),
	}

	for _, t := range tags {
		checks = append(checks, resource.TestCheckTypeSetElemAttr("maas_node_script.test", "tags.*", t))
	}

	for _, fw := range forHardware {
		checks = append(checks, resource.TestCheckTypeSetElemAttr("maas_node_script.test", "for_hardware.*", fw))
	}

	updatedChecks := []resource.TestCheckFunc{
		testAccMAASNodeScriptCheckExists("maas_node_script.test", &nodeScript),
		resource.TestCheckResourceAttr("maas_node_script.test", "script", encodedScript),
		resource.TestCheckResourceAttr("maas_node_script.test", "name", name),
		resource.TestCheckResourceAttr("maas_node_script.test", "script_type", updatedScriptType),
		resource.TestCheckResourceAttr("maas_node_script.test", "title", updatedTitle),
		resource.TestCheckResourceAttr("maas_node_script.test", "description", updatedDescription),
		resource.TestCheckResourceAttr("maas_node_script.test", "comment", updatedComment),
		resource.TestCheckResourceAttr("maas_node_script.test", "parallel", updatedParallel),
		resource.TestCheckResourceAttr("maas_node_script.test", "timeout", updatedTimeout),
		resource.TestCheckResourceAttr("maas_node_script.test", "hardware_type", updatedHardwareType),
		resource.TestCheckResourceAttr("maas_node_script.test", "apply_configured_networking", fmt.Sprintf("%t", updatedApplyConfiguredNetworking)),
		resource.TestCheckResourceAttr("maas_node_script.test", "destructive", fmt.Sprintf("%t", updatedDestructive)),
		resource.TestCheckResourceAttr("maas_node_script.test", "may_reboot", fmt.Sprintf("%t", updatedMayReboot)),
		resource.TestCheckResourceAttr("maas_node_script.test", "recommission", fmt.Sprintf("%t", updatedRecommission)),
		resource.TestCheckResourceAttr("maas_node_script.test", "tags.#", fmt.Sprintf("%v", len(updatedTags))),
		resource.TestCheckResourceAttr("maas_node_script.test", "for_hardware.#", fmt.Sprintf("%v", len(updatedForHardware))),
	}

	for _, t := range updatedTags {
		updatedChecks = append(updatedChecks, resource.TestCheckTypeSetElemAttr("maas_node_script.test", "tags.*", t))
	}

	for _, fw := range updatedForHardware {
		updatedChecks = append(updatedChecks, resource.TestCheckTypeSetElemAttr("maas_node_script.test", "for_hardware.*", fw))
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testutils.PreCheck(t, nil) },
		Providers:    testutils.TestAccProviders,
		CheckDestroy: testAccCheckMAASNodeScriptDestroy,
		ErrorCheck:   func(err error) error { return err },
		Steps: []resource.TestStep{
			{
				Config: testAccMAASNodeScript(
					scriptType, name, title, description, comment, parallel, timeout, hardwareType,
					applyConfiguredNetworking, destructive, mayReboot, recommission,
					forHardware, tags,
					packages,
				),
				Check: resource.ComposeTestCheckFunc(checks...),
			},
			// Test update script
			{
				Config: testAccMAASNodeScript(
					updatedScriptType, name, updatedTitle, updatedDescription, updatedComment, updatedParallel, updatedTimeout, updatedScriptType,
					updatedApplyConfiguredNetworking, updatedDestructive, updatedMayReboot, updatedRecommission,
					updatedForHardware, updatedTags,
					updatedPackages,
				),
				Check: resource.ComposeTestCheckFunc(checks...),
			},
			// Test import using name
			{
				ResourceName: "maas_node_script.test",
				ImportState:  true,
				ImportStateCheck: func(is []*terraform.InstanceState) error {
					var nodeScript *terraform.InstanceState
					if len(is) != 1 {
						return fmt.Errorf("expected 1 state: %#v", t)
					}
					nodeScript = is[0]
					assert.Equal(t, nodeScript.Attributes["name"], name)
					assert.Equal(t, nodeScript.Attributes["description"], description)
					return nil
				},
			},
			// Test import using ID
			{
				ResourceName:      "maas_node_script.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccMAASNodeScriptCheckExists(rn string, nodeScript *entity.NodeScript) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rn]
		if !ok {
			return fmt.Errorf("resource not found: %s\n %#v", rn, s.RootModule().Resources)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource id not set")
		}

		conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

		gotNodeScript, err := getNodeScript(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error getting nodeScript: %s", err)
		}

		*nodeScript = *gotNodeScript

		return nil
	}
}

func testAccMAASNodeScript(
	scriptType, name, title, description, comment, parallel, timeout, hardwareType string,
	applyConfiguredNetworking, destructive, mayReboot, recommission bool,
	forHardware, tags []string,
	packages map[string][]string,
) string {
	packagesMapString, _ := json.Marshal(packages)

	return fmt.Sprintf(`
resource "maas_node_script" "test" {
  script                      = file("${path.module}/scripts/dummy.sh")
  script_type                 = %v
  name                        = %v
  title                       = %v
  description                 = %v
  comment                     = %v
  parallel                    = %v
  timeout                     = %v
  hardware_type               = %v
  for_hardware                = %v
  packages                    = %v
  apply_configured_networking = %v
  destructive                 = %v
  may_reboot                  = %v
  recommission                = %v
  tags                        = %v
}
`,
		scriptType, name, title, description, comment, parallel, timeout, hardwareType,
		testutils.StringifySliceAsLiteralArray(forHardware), packagesMapString, applyConfiguredNetworking,
		destructive, mayReboot, recommission, testutils.StringifySliceAsLiteralArray(tags),
	)
}

func testAccCheckMAASNodeScriptDestroy(s *terraform.State) error {
	// retrieve the connection established in Provider configuration
	conn := testutils.TestAccProvider.Meta().(*maas.ClientConfig).Client

	// loop through the resources in state, verifying each maas_node_script
	// is destroyed
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "maas_node_script" {
			continue
		}

		// Retrieve our maas_node_script by referencing it's state ID for API lookup
		response, err := getNodeScript(conn, rs.Primary.ID)
		if err == nil {
			if response != nil && fmt.Sprintf("%v", response.ID) == rs.Primary.ID {
				return fmt.Errorf("MAAS NodeScript (%s) still exists.", rs.Primary.ID)
			}

			return nil
		}

		// If the error is equivalent to 404 not found, the maas_node_script is destroyed.
		// Otherwise return the error
		if !strings.Contains(err.Error(), "404 Not Found") {
			return err
		}
	}

	return nil
}

func getNodeScript(client *client.Client, identifier string) (*entity.NodeScript, error) {
	nodeScripts, err := client.NodeScripts.Get(&entity.NodeScriptReadParams{})
	if err != nil {
		return nil, err
	}

	for _, s := range nodeScripts {
		if fmt.Sprintf("%v", s.ID) == identifier || s.Name == identifier {
			return &s, nil
		}
	}

	return nil, fmt.Errorf("404 Not Found: %v", identifier)
}
