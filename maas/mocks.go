package maas

import (
	"github.com/juju/collections/set"
	"github.com/juju/gomaasapi"
)

type MockOwnerDataHolder struct {
}

func (m *MockOwnerDataHolder) OwnerData() map[string]string {
	return nil
}
func (m *MockOwnerDataHolder) SetOwnerData(map[string]string) error {
	return nil
}

type MockMachine struct {
	MockOwnerDataHolder
	systemId string
}

func (m *MockMachine) SystemID() string {
	return m.systemId
}
func (m *MockMachine) Hostname() string {
	return ""
}
func (m *MockMachine) FQDN() string {
	return ""
}
func (m *MockMachine) Tags() []string {
	return nil
}
func (m *MockMachine) OperatingSystem() string {
	return ""
}
func (m *MockMachine) DistroSeries() string {
	return ""
}
func (m *MockMachine) Architecture() string {
	return ""
}
func (m *MockMachine) Memory() int {
	return 0
}
func (m *MockMachine) CPUCount() int {
	return 0
}
func (m *MockMachine) IPAddresses() []string {
	return nil
}
func (m *MockMachine) PowerState() string {
	return ""
}
func (m *MockMachine) Devices(gomaasapi.DevicesArgs) ([]gomaasapi.Device, error) {
	return nil, nil
}
func (m *MockMachine) StatusName() string {
	return ""
}
func (m *MockMachine) StatusMessage() string {
	return ""
}
func (m *MockMachine) BootInterface() gomaasapi.Interface {
	return nil
}
func (m *MockMachine) InterfaceSet() []gomaasapi.Interface {
	return nil
}
func (m *MockMachine) Interface(id int) gomaasapi.Interface {
	return nil
}
func (m *MockMachine) PhysicalBlockDevices() []gomaasapi.BlockDevice {
	return nil
}
func (m *MockMachine) PhysicalBlockDevice(id int) gomaasapi.BlockDevice {
	return nil
}
func (m *MockMachine) BlockDevices() []gomaasapi.BlockDevice {
	return nil
}
func (m *MockMachine) BlockDevice(id int) gomaasapi.BlockDevice {
	return nil
}
func (m *MockMachine) Partition(id int) gomaasapi.Partition {
	return nil
}
func (m *MockMachine) Zone() gomaasapi.Zone {
	return nil
}
func (m *MockMachine) Pool() gomaasapi.Pool {
	return nil
}
func (m *MockMachine) Start(gomaasapi.StartArgs) error {
	return nil
}
func (m *MockMachine) CreateDevice(gomaasapi.CreateMachineDeviceArgs) (gomaasapi.Device, error) {
	return nil, nil
}

type MockController struct {
	machines []MockMachine
}

func (m *MockController) Capabilities() set.Strings {
	return set.NewStrings()
}
func (m *MockController) BootResources() ([]gomaasapi.BootResource, error) {
	return nil, nil
}
func (m *MockController) Fabrics() ([]gomaasapi.Fabric, error) {
	return nil, nil
}
func (m *MockController) Spaces() ([]gomaasapi.Space, error) {
	return nil, nil
}
func (m *MockController) StaticRoutes() ([]gomaasapi.StaticRoute, error) {
	return nil, nil
}
func (m *MockController) Zones() ([]gomaasapi.Zone, error) {
	return nil, nil
}
func (m *MockController) Pools() ([]gomaasapi.Pool, error) {
	return nil, nil
}
func (m *MockController) Machines(machineArgs gomaasapi.MachinesArgs) ([]gomaasapi.Machine, error) {
	var result []gomaasapi.Machine
	for _, machine := range m.machines {
		for _, id := range machineArgs.SystemIDs {
			if machine.SystemID() == id {
				ma := machine
				result = append(result, &ma)
			}
		}
	}
	return result, nil
}
func (m *MockController) AllocateMachine(gomaasapi.AllocateMachineArgs) (gomaasapi.Machine, gomaasapi.ConstraintMatches, error) {
	var c gomaasapi.ConstraintMatches
	return nil, c, nil
}
func (m *MockController) ReleaseMachines(gomaasapi.ReleaseMachinesArgs) error {
	return nil
}
func (m *MockController) Devices(gomaasapi.DevicesArgs) ([]gomaasapi.Device, error) {
	return nil, nil
}
func (m *MockController) CreateDevice(gomaasapi.CreateDeviceArgs) (gomaasapi.Device, error) {
	return nil, nil
}
func (m *MockController) Files(prefix string) ([]gomaasapi.File, error) {
	return nil, nil
}
func (m *MockController) GetFile(filename string) (gomaasapi.File, error) {
	return nil, nil
}
func (m *MockController) AddFile(gomaasapi.AddFileArgs) error {
	return nil
}
func (m *MockController) Domains() ([]gomaasapi.Domain, error) {
	return nil, nil
}
