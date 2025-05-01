resource "maas_vm_host" "kvm" {
  type          = "virsh"
  power_address = "qemu+ssh://ubuntu@10.113.1.24/system"
  tags = [
    "pod-console-logging",
    "virtual",
    "kvm",
  ]
}

resource "maas_vm_host" "lxd_password" {
  type          = "lxd"
  power_address = "10.10.0.1"
  lxd_project   = "test-project"
  password      = "my-password"
}

resource "maas_vm_host" "lxd_certificate" {
  type          = "lxd"
  power_address = "10.10.0.1"
  lxd_project   = "test-project"
  certificate   = "-----BEGIN CERTIFICATE-----\n certificate-goes-here =\n-----END CERTIFICATE-----\n"
  key           = "-----BEGIN PRIVATE KEY-----\n key-goes-here ==\n-----END PRIVATE KEY-----\n"
}

  