package maas

import (
	"context"
	"fmt"
	mock_gomaasapi "terraform-provider-maas/test/mocks"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/juju/gomaasapi"
	"github.com/stretchr/testify/assert"
)

func TestDeployMaasMachine(t *testing.T) {
	stateId := "test-id"

	machineMockFuncMap := map[string]func(*gomock.Controller) *mock_gomaasapi.MockMachine{
		"machineStartErr": func(ctrl *gomock.Controller) *mock_gomaasapi.MockMachine {
			machine := mock_gomaasapi.NewMockMachine(ctrl)
			machine.
				EXPECT().
				SystemID().
				Return(stateId)
			machine.
				EXPECT().
				Start(gomaasapi.StartArgs{}).
				Return(fmt.Errorf("start error"))
			return machine
		},
	}

	testCases := []struct {
		name            string
		machineMockFunc func(*gomock.Controller) *mock_gomaasapi.MockMachine
		err             error
	}{
		{
			name:            "machine is not found",
			machineMockFunc: nil,
			err:             fmt.Errorf("failed to get machine (test-id): machine (test-id) was not found"),
		},
		{
			name:            "machine start error",
			machineMockFunc: machineMockFuncMap["machineStartErr"],
			err:             fmt.Errorf("failed to start machine (test-id): start error"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			d := schema.ResourceData{}
			d.SetId(stateId)

			mockCtrl := gomock.NewController(t)

			var machines []gomaasapi.Machine
			if testCase.machineMockFunc != nil {
				machines = append(machines, testCase.machineMockFunc(mockCtrl))
			}
			client := mock_gomaasapi.NewMockController(mockCtrl)
			client.
				EXPECT().
				Machines(gomock.Eq(gomaasapi.MachinesArgs{SystemIDs: []string{d.Id()}})).
				Return(machines, nil)

			err := deployMaasMachine(&d, client, context.Background())

			assert.Equal(t, testCase.err, err)
		})
	}
}

func TestGetMaasMachineStatusFunc(t *testing.T) {
	machineId := "test-id"

	machineMockFuncMap := map[string]func(*gomock.Controller) *mock_gomaasapi.MockMachine{
		"machineDeployed": func(ctrl *gomock.Controller) *mock_gomaasapi.MockMachine {
			machine := mock_gomaasapi.NewMockMachine(ctrl)
			machine.
				EXPECT().
				StatusName().
				Return("Deployed").
				Times(2)
			return machine
		},
	}

	testCases := []struct {
		name            string
		machineMockFunc func(*gomock.Controller) *mock_gomaasapi.MockMachine
		status          string
		err             error
	}{
		{
			name:            "machine is not found",
			machineMockFunc: nil,
			status:          "",
			err:             fmt.Errorf("machine (%s) was not found", machineId),
		},
		{
			name:            "machine is deployed",
			machineMockFunc: machineMockFuncMap["machineDeployed"],
			status:          "Deployed",
			err:             nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mockCtrl := gomock.NewController(t)

			var machines []gomaasapi.Machine
			if testCase.machineMockFunc != nil {
				machines = append(machines, testCase.machineMockFunc(mockCtrl))
			}
			client := mock_gomaasapi.NewMockController(mockCtrl)
			client.
				EXPECT().
				Machines(gomock.Eq(gomaasapi.MachinesArgs{SystemIDs: []string{machineId}})).
				Return(machines, nil)

			machine, status, err := getMaasMachineStatusFunc(client, machineId)()

			if testCase.err != nil {
				assert.Equal(t, testCase.err, err)
			} else {
				if assert.Nil(t, err) {
					assert.Equal(t, testCase.status, status)
					assert.Equal(t, machine, machines[0])
				}
			}
		})
	}
}
