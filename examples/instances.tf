resource "maas_instance" "kvm" {
  count = 2
  allocate_hostname = maas_vm_host_machine.kvm[count.index].hostname
  allocate_min_cpu_count = 1
  allocate_min_memory = 2048
  allocate_tags = [
    maas_tag.virtual.name,
    maas_tag.kvm.name,
    maas_tag.ubuntu.name,
  ]
  deploy_distro_series = "focal"
}
