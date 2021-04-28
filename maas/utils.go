package maas

import (
	"encoding/base64"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/ionutbalutoiu/gomaasclient/gmaw"
	"github.com/ionutbalutoiu/gomaasclient/maas"
	"github.com/juju/gomaasapi"
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

func getMachineStatusFunc(client *gomaasapi.MAASObject, systemId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		machineManager, err := maas.NewMachineManager(systemId, gmaw.NewMachine(client))
		if err != nil {
			log.Printf("[ERROR] Unable to get machine (%s) status: %s\n", systemId, err)
			return nil, "", err
		}
		machine := machineManager.Current()

		log.Printf("[DEBUG] Machine (%s) status: %s\n", systemId, machine.StatusName)
		return machine, machine.StatusName, nil
	}
}
