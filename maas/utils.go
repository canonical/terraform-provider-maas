package maas

import (
	"encoding/base64"

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

func getMaasMachine(client gomaasapi.Controller, systemId string) (gomaasapi.Machine, error) {
	machines, err := client.Machines(gomaasapi.MachinesArgs{SystemIDs: []string{systemId}})
	if err != nil {
		return nil, err
	}

	return machines[0], nil
}
