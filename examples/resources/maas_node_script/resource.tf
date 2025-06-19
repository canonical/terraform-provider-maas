resource "maas_node_script" "tf_node_script" {
  script = base64encode(file("${path.module}/scripts/dummy.sh"))
}
