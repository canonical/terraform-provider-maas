package maas

import (
	"encoding/base64"
	"fmt"
	"net/mail"
	"strings"

	"github.com/canonical/gomaasclient/client"
	"github.com/canonical/gomaasclient/entity"
	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/go-cty/cty/gocty"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func base64Encode(data []byte) string {
	if isBase64Encoded(data) {
		return string(data)
	}

	return base64.StdEncoding.EncodeToString(data)
}

func isBase64Encoded(data []byte) bool {
	_, err := base64.StdEncoding.DecodeString(string(data))
	return err == nil
}

func convertToStringSlice(field interface{}) []string {
	if field == nil {
		return nil
	}
	fieldSlice := field.([]interface{})
	result := make([]string, len(fieldSlice))
	for i, value := range fieldSlice {
		result[i] = value.(string)
	}
	return result
}

func isElementIPAddress(i interface{}, p cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	attr := p[len(p)-1].(cty.IndexStep)
	var index int
	if err := gocty.FromCtyValue(attr.Key, &index); err != nil {
		return diag.FromErr(err)
	}
	ws, es := validation.IsIPAddress(i, fmt.Sprintf("element %v", index))

	for _, w := range ws {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Warning,
			Summary:       w,
			AttributePath: p,
		})
	}
	for _, e := range es {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       e.Error(),
			AttributePath: p,
		})
	}
	return diags
}

func isEmailAddress(i interface{}, p cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	attr := p[len(p)-1].(cty.GetAttrStep)

	v, ok := i.(string)
	if !ok {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       fmt.Sprintf("expected type of %q to be string", attr.Name),
			AttributePath: p,
		})
	}

	if _, err := mail.ParseAddress(i.(string)); err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity:      diag.Error,
			Summary:       fmt.Sprintf("expected %s to be a valid e-mail address, got: %s", attr.Name, v),
			AttributePath: p,
		})
	}

	return diags
}

func getNetworkInterface(client *client.Client, machineSystemID string, identifier string) (*entity.NetworkInterface, error) {
	networkInterfaces, err := client.NetworkInterfaces.Get(machineSystemID)
	if err != nil {
		return nil, err
	}
	for _, n := range networkInterfaces {
		if n.MACAddress == identifier || n.Name == identifier || fmt.Sprintf("%v", n.ID) == identifier {
			return &n, nil
		}
	}
	return nil, fmt.Errorf("network interface (%s) was not found on machine (%s)", identifier, machineSystemID)
}

func setTerraformState(d *schema.ResourceData, tfState map[string]interface{}) error {
	if val, ok := tfState["id"]; ok {
		d.SetId(val.(string))
		delete(tfState, "id")
	}
	for k, v := range tfState {
		// NOTE: Ignore R001. We have this method being invoked in multiple places,
		// however key values are actually string literals. Consider this a false positive.
		// https://github.com/bflad/tfproviderlint/tree/main/passes/R001
		//
		//lintignore:R001
		if err := d.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Get the system ID of the relevant entity from resource data that accepts either a `machine` or `device`.
func getMachineOrDeviceSystemID(client *client.Client, d *schema.ResourceData) (string, error) {
	if d.Get("machine") != "" {
		machine, err := getMachine(client, d.Get("machine").(string))
		if err != nil && !strings.Contains(err.Error(), "404 Not Found") {
			return "", err
		}
		return machine.SystemID, nil
	}

	if d.Get("device") != "" {
		device, err := getDevice(client, d.Get("device").(string))
		if err != nil && !strings.Contains(err.Error(), "404 Not Found") {
			return "", err
		}
		return device.SystemID, nil
	}
	return "", fmt.Errorf("either `machine` or `device` is required")
}
