package maas

import (
	"github.com/juju/gomaasapi"
)

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

func getMaasMachine(client gomaasapi.Controller, systemId string) (gomaasapi.Machine, error) {
	machines, err := client.Machines(gomaasapi.MachinesArgs{SystemIDs: []string{systemId}})
	if err != nil {
		return nil, err
	}

	return machines[0], nil
}
