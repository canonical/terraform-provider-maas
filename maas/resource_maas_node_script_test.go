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

	name := acctest.RandomWithPrefix("tf-node-script")

	scriptType := "commissioning"
	title := "Initial Title"
	description := "initial description"
	parallel := "instance"
	timeout := "0:10:00"
	hardwareType := "node"
	applyConfiguredNetworking := true
	destructive := true
	mayReboot := true
	recommission := true
	forHardware := []string{"system_vendor:canonical", "system_product:maas", "pci:1234:5678"}
	tags := []string{"dummy", "script", "destructive", hardwareType}
	packages := map[string][]string{"snap": {"maas", "maas-test-db"}}
	packagesBytes, _ := json.Marshal(packages)
	parameters := map[string]map[string]string{"storage": {"type": "string"}}
	parametersBytes, _ := json.Marshal(parameters)
	results := map[string]map[string]string{"badblocks": {"title": "Bad blocks"}}
	resultsBytes, _ := json.Marshal(results)

	scriptRaw := testAccMAASNodeScriptWithMetadata(
		scriptType, name, title, description, parallel, timeout, hardwareType,
		applyConfiguredNetworking, destructive, mayReboot, recommission,
		forHardware, tags,
		packages,
		parameters, results,
	)
	encodedScript := base64.StdEncoding.EncodeToString([]byte(scriptRaw))

	updatedScriptType := "testing"
	updatedTitle := "Updated Title"
	updatedDescription := "updated description"
	updatedParallel := "any"
	updatedHardwareType := "storage"
	updatedApplyConfiguredNetworking := false
	updatedDestructive := false
	updatedMayReboot := false
	updatedRecommission := false
	updatedForHardware := []string{"system_vendor:canonical", "system_product:maas-testdb"}
	updatedTags := []string{"dummy", "script", updatedHardwareType}
	updatedPackages := map[string][]string{"snap": {"maas", "maas-test-db", "core24"}}
	updatedPackagesBytes, _ := json.Marshal(updatedPackages)
	updatedParameters := map[string]map[string]string{"bytes": {"type": "int"}}
	updatedParametersBytes, _ := json.Marshal(updatedParameters)
	updatedResults := map[string]map[string]string{"goodblocks": {"title": "Good blocks"}}
	updatedResultsBytes, _ := json.Marshal(updatedResults)

	updatedScriptRaw := testAccMAASNodeScriptWithMetadata(
		updatedScriptType, name, updatedTitle, updatedDescription, updatedParallel, timeout, updatedHardwareType,
		updatedApplyConfiguredNetworking, updatedDestructive, updatedMayReboot, updatedRecommission,
		updatedForHardware, updatedTags,
		updatedPackages,
		updatedParameters, updatedResults,
	)
	encodedUpdatedScript := base64.StdEncoding.EncodeToString([]byte(updatedScriptRaw))

	checks := []resource.TestCheckFunc{
		testAccMAASNodeScriptCheckExists("maas_node_script.test", &nodeScript),
		resource.TestCheckResourceAttr("maas_node_script.test", "script", encodedScript),
		resource.TestCheckResourceAttr("maas_node_script.test", "name", name),
		resource.TestCheckResourceAttr("maas_node_script.test", "script_type", scriptType),
		resource.TestCheckResourceAttr("maas_node_script.test", "title", title),
		resource.TestCheckResourceAttr("maas_node_script.test", "description", description),
		resource.TestCheckResourceAttr("maas_node_script.test", "parallel", parallel),
		resource.TestCheckResourceAttr("maas_node_script.test", "timeout", timeout),
		resource.TestCheckResourceAttr("maas_node_script.test", "hardware_type", hardwareType),
		resource.TestCheckResourceAttr("maas_node_script.test", "apply_configured_networking", fmt.Sprintf("%t", applyConfiguredNetworking)),
		resource.TestCheckResourceAttr("maas_node_script.test", "destructive", fmt.Sprintf("%t", destructive)),
		resource.TestCheckResourceAttr("maas_node_script.test", "may_reboot", fmt.Sprintf("%t", mayReboot)),
		resource.TestCheckResourceAttr("maas_node_script.test", "recommission", fmt.Sprintf("%t", recommission)),
		resource.TestCheckResourceAttr("maas_node_script.test", "packages", string(packagesBytes)),
		resource.TestCheckResourceAttr("maas_node_script.test", "parameters", string(parametersBytes)),
		resource.TestCheckResourceAttr("maas_node_script.test", "results", string(resultsBytes)),
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
		resource.TestCheckResourceAttr("maas_node_script.test", "script", encodedUpdatedScript),
		resource.TestCheckResourceAttr("maas_node_script.test", "name", name),
		resource.TestCheckResourceAttr("maas_node_script.test", "script_type", updatedScriptType),
		resource.TestCheckResourceAttr("maas_node_script.test", "title", updatedTitle),
		resource.TestCheckResourceAttr("maas_node_script.test", "description", updatedDescription),
		resource.TestCheckResourceAttr("maas_node_script.test", "parallel", updatedParallel),
		resource.TestCheckResourceAttr("maas_node_script.test", "timeout", timeout),
		resource.TestCheckResourceAttr("maas_node_script.test", "hardware_type", updatedHardwareType),
		resource.TestCheckResourceAttr("maas_node_script.test", "apply_configured_networking", fmt.Sprintf("%t", updatedApplyConfiguredNetworking)),
		resource.TestCheckResourceAttr("maas_node_script.test", "destructive", fmt.Sprintf("%t", updatedDestructive)),
		resource.TestCheckResourceAttr("maas_node_script.test", "may_reboot", fmt.Sprintf("%t", updatedMayReboot)),
		resource.TestCheckResourceAttr("maas_node_script.test", "recommission", fmt.Sprintf("%t", updatedRecommission)),
		resource.TestCheckResourceAttr("maas_node_script.test", "packages", string(updatedPackagesBytes)),
		resource.TestCheckResourceAttr("maas_node_script.test", "parameters", string(updatedParametersBytes)),
		resource.TestCheckResourceAttr("maas_node_script.test", "results", string(updatedResultsBytes)),
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
				Config: testAccMAASNodeScript(encodedScript),
				Check:  resource.ComposeTestCheckFunc(checks...),
			},
			// Test update script
			{
				Config: testAccMAASNodeScript(encodedUpdatedScript),
				Check:  resource.ComposeTestCheckFunc(updatedChecks...),
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

func testAccMAASNodeScriptWithMetadata(
	scriptType, name, title, description, parallel, timeout, hardwareType string,
	applyConfiguredNetworking, destructive, mayReboot, recommission bool,
	forHardware, tags []string,
	packages map[string][]string,
	parameters, results map[string]map[string]string,
) string {
	packagesMapString, _ := json.Marshal(packages)
	parametersMapString, _ := json.Marshal(parameters)
	resultsMapString, _ := json.Marshal(results)

	return fmt.Sprintf(`#!/bin/bash

# --- Start MAAS 1.0 script metadata ---
# script_type: %q
# name: %q
# title: %q
# description: %q
# parallel: %q
# timeout: %q
# hardware_type: %q
# for_hardware: %v
# packages: %v
# apply_configured_networking: %v
# destructive: %v
# may_reboot: %v
# recommission: %v
# tags: %v
# parameters: %v
# results: %v
# --- End MAAS 1.0 script metadata ---

echo "Hello World"
`,
		scriptType, name, title, description, parallel, timeout, hardwareType,
		testutils.StringifySliceAsLiteralArray(forHardware), string(packagesMapString), applyConfiguredNetworking,
		destructive, mayReboot, recommission, testutils.StringifySliceAsLiteralArray(tags), string(parametersMapString),
		string(resultsMapString),
	)
}

func testAccMAASNodeScript(script string) string {
	return fmt.Sprintf(`
resource "maas_node_script" "test" {
  script = %q
}
`, script)
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
