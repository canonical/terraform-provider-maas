package maas

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
)

func testResourceMaasMachineInstanceStateDataV0() map[string]any {
	return map[string]any{
		"power_parameters": map[string]any{
			"power_user": "ubuntu",
		},
	}
}

func testResourceMaasMachineInstanceStateDataV1() map[string]any {
	flattenedV0, _ := structure.FlattenJsonToString(map[string]any{
		"power_user": "ubuntu",
	})

	return map[string]any{"power_parameters": flattenedV0}
}

func TestResourceMaasMachineInstanceStateUpgradeV0(t *testing.T) {
	ctx := context.Background()
	expected := testResourceMaasMachineInstanceStateDataV1()

	actual, err := resourceMaasMachineStateUpgradeV0(ctx, testResourceMaasMachineInstanceStateDataV0(), nil)
	if err != nil {
		t.Fatalf("error migrating state: %s", err)
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatalf("\n\nexpected:\n\n%#v\n\ngot:\n\n%#v\n\n", expected, actual)
	}
}
