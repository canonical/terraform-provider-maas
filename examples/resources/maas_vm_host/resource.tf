resource "maas_vm_host" "kvm_virsh" {
  type          = "virsh"
  power_address = "qemu+ssh://ubuntu@10.113.1.24/system"
  tags = [
    "pod-console-logging",
    "virtual",
    "kvm",
  ]
}

# A LXD VM host. A certificate will be generated if not provided
resource "maas_vm_host" "lxd_no_certificate" {
  type          = "lxd"
  power_address = "10.10.0.1"
  project       = "test-project"
  password      = "my-lxd-trust-password"
}

# New, untrusted certificates can be trusted by specifying the LXD trust password or token
resource "maas_vm_host" "lxd_new_certificate" {
  type          = "lxd"
  power_address = "10.10.0.1"
  project       = "test-project"
  password      = "my-lxd-trust-password"
  certificate   = "-----BEGIN CERTIFICATE-----\n certificate-goes-here =\n-----END CERTIFICATE-----\n"
  key           = "-----BEGIN PRIVATE KEY-----\n key-goes-here ==\n-----END PRIVATE KEY-----\n"
}

# If the certificate is already trusted by your LXD cluster, the trust password can be omitted
resource "maas_vm_host" "lxd_pre_trusted_certificate" {
  type          = "lxd"
  power_address = "10.10.0.1"
  project       = "test-project"
  certificate   = "-----BEGIN CERTIFICATE-----\n certificate-goes-here =\n-----END CERTIFICATE-----\n"
  key           = "-----BEGIN PRIVATE KEY-----\n key-goes-here ==\n-----END PRIVATE KEY-----\n"
}
