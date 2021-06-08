resource "maas_instance" "kvm" {
  count = 2
  allocate_params {
    hostname = maas_vm_host_machine.kvm[count.index].hostname
    min_cpu_count = 1
    min_memory = 2048
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

resource "maas_instance" "machine_05" {
  allocate_params {
    hostname = "machine-05"
  }
  deploy_params {
    distro_series = "focal"
  }
  network_interfaces {
    name = "enp1s0"
    subnet_cidr = "10.99.0.0/16"
    ip_address = "10.99.123.123"
  }
  network_interfaces {
    name = "enp2s0"
    subnet_cidr = "10.10.0.0/16"
  }
  network_interfaces {
    // It will mark the interface as disconnected
    name = "enp3s0"
  }
}
