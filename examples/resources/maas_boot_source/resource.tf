resource "maas_boot_source" "test_boot_source" {
  url              = "http://images.maas.io/ephemeral-v3/candidate/"
  keyring_filename = "/snap/maas/current/usr/share/keyrings/ubuntu-cloudimage-keyring.gpg"
}
