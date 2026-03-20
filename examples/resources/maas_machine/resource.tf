# Add an existing Virsh VM
resource "maas_machine" "virsh_vm1" {
  power_type = "virsh"
  power_parameters = jsonencode({
    power_address = "qemu+ssh://ubuntu@10.113.1.26/system"
    power_id      = "test-vm1"
  })
  pxe_mac_address = "52:54:00:89:f5:3e"
}

# Add an exisiting LXD VM, specifying a commissioning script and its parameters.
# This might be applicable to a node-script with the 'noauto' tag, in order to run this
# solely for a particular machine.
resource "maas_node_script" "my-script" {
  script = base64encode(<<-EOF
#!/bin/bash

# --- Start MAAS 1.0 script metadata ---
# name: commissioning-test
# title: Testing out commissioning
# description: This is an example commissioning script that simply echoes a message.
# script_type: commissioning
# tags: noauto
# parameters:
#   msg:
#     type: string
#     required: true
#     argument_format: '{input}'
#     description: Message to echo.
# --- End MAAS 1.0 script metadata ---

echo "msg found: $1"
EOF
  )
}

resource "maas_machine" "myvm" {
  hostname        = "my-special-machine"
  architecture    = "amd64/generic"
  power_type      = "lxd"
  pxe_mac_address = "00:16:3e:f9:8e:bb"
  power_parameters = jsonencode({
    project       = "default",
    certificate   = file("cert.pem")
    key           = file("pass.key")
    power_address = "10.10.0.1",
    instance_name = "test-machine",
  })
  script_parameters = {
    commissioning-test_msg = "hello"
  }
  commissioning_scripts = [maas_node_script.my-script.name]
}

# Add a DPU machine, which will be treated as a DPU by MAAS and have the appropriate hardware type set.
# MAAS version >= 3.6 is required for this.
resource "maas_machine" "dpu1" {
  hostname        = "my-dpu-machine"
  architecture    = "arm64/generic"
  min_hwe_kernel  = "hwe-22.04"
  power_type      = "redfish"
  pxe_mac_address = "00:11:22:33:44:55"
  power_parameters = jsonencode({
    power_address = "10.10.10.42",
    power_user    = "admin",
    power_pass    = "admin",
  })
  is_dpu = true
}

# Add an IPMI and non-IPMI machine, note the optionality of pxe_mac_address *only* for the IPMI power type

resource "maas_machine" "non_ipmi_machine" {
  power_type   = "redfish"
  architecture = "amd64/generic"
  power_parameters = jsonencode({
    power_address = "10.10.10.26"
    power_user    = "admin"
    power_pass    = "password"
  })
  hostname        = "redfishTestMachine"
  pxe_mac_address = "00:11:22:33:44:55"
}

resource "maas_machine" "ipmi_machine" {
  power_type   = "ipmi"
  architecture = "amd64/generic"
  power_parameters = jsonencode({
    power_address = "10.10.10.26"
    power_user    = "admin"
    power_pass    = "password"
  })
  hostname = "ipmiTestMachineNoPxe"
}
