resource "maas_instance" "kvm" {
  count = 2
  allocate_params {
    hostname      = maas_vm_host_machine.kvm[count.index].hostname
    min_cpu_count = 1
    min_memory    = 2048
    tags = [
      maas_tag.virtual.name,
      maas_tag.kvm.name,
      maas_tag.ubuntu.name,
    ]
  }
  deploy_params {
    distro_series = "focal"
  }
}
