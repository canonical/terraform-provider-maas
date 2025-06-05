resource "maas_node_script" "tf_node_script" {
  script      = base64encode(file("${path.module}/scripts/dummy.sh"))
  script_type = "commissioning"

  name        = "dummy-script"
  title       = "Dummy Script"
  description = "A dummy node script to demonstrate the Terraform resource"
  comment     = "First attempt to create a node script"

  parallel = "instance"
  timeout  = "00:20:00"

  hardware_type = "node"
  for_hardware = [
    "system_vendor:canonical",
    "system_product:maas",
  ]

  packages = jsonencode({
    snap = ["maas", "maas-test-db"]
    apt  = ["bind9"]
  })

  apply_configured_networking = true
  destructive                 = true
  may_reboot                  = true
  recommission                = true

  tags = [
    "dummy",
    "script",
  ]
}
